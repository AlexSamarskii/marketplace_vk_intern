package entity

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	govalidator "github.com/asaskevich/govalidator"
)

type Advertisement struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" valid:"required,length(3|50)"`
	Description string    `json:"description" valid:"required,length(10|500)"`
	ImageURL    string    `json:"image_url" valid:"required,url,imgext"`
	Price       float64   `json:"price" valid:"required,float"`
	UserID      int       `json:"user_id" valid:"required"`
	AuthorLogin string    `json:"author_login"`
	IsMine      bool      `json:"is_mine"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	AdTitleMinLen       = 3
	AdTitleMaxLen       = 50
	AdDescriptionMinLen = 10
	AdDescriptionMaxLen = 500
	AdImageURLMaxLen    = 2048
	AdPriceMin          = 0.0
	AdPriceMax          = 1_000_000_000
	AdImageMaxBytes     = 5 << 20 // 5 MB
	AdImageMaxWidth     = 4096
	AdImageMaxHeight    = 4096
)

var allowedImageExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
}

func init() {
	govalidator.TagMap["imgext"] = govalidator.Validator(func(str string) bool {
		u, err := url.Parse(str)
		if err != nil {
			return false
		}
		ext := strings.ToLower(path.Ext(u.Path))
		_, ok := allowedImageExt[ext]
		return ok
	})
}

type FieldErrors map[string]string

type AdvValidationError struct {
	Fields FieldErrors
}

func (e *AdvValidationError) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("validation failed")
	for f, m := range e.Fields {
		b.WriteString("; ")
		b.WriteString(f)
		b.WriteString(": ")
		b.WriteString(m)
	}
	return b.String()
}

func parseGoValidatorErr(err error) FieldErrors {
	if err == nil {
		return nil
	}
	m := govalidator.ErrorsByField(err)
	fe := make(FieldErrors, len(m))
	for k, v := range m {
		fe[k] = v
	}
	return fe
}

func (a *Advertisement) Validate() (bool, error) {
	govalidator.SetFieldsRequiredByDefault(false)

	_, err := govalidator.ValidateStruct(a)

	fe := parseGoValidatorErr(err)
	if fe == nil {
		fe = FieldErrors{}
	}

	l := len(strings.TrimSpace(a.Title))
	if l < AdTitleMinLen || l > AdTitleMaxLen {
		fe["title"] = fmt.Sprintf("длина должна быть от %d до %d", AdTitleMinLen, AdTitleMaxLen)
	}

	l = len(strings.TrimSpace(a.Description))
	if l < AdDescriptionMinLen || l > AdDescriptionMaxLen {
		fe["description"] = fmt.Sprintf("длина должна быть от %d до %d", AdDescriptionMinLen, AdDescriptionMaxLen)
	}

	if e := validatePriceRange(a.Price); e != nil {
		fe["price"] = e.Error()
	}

	if e := validateImageURLBasic(a.ImageURL); e != nil {
		fe["image_url"] = e.Error()
	}

	if a.UserID <= 0 {
		fe["user_id"] = "должен быть > 0"
	}

	if len(fe) > 0 {
		return false, &AdvValidationError{Fields: fe}
	}
	return true, nil
}

// validatePriceRange проверяет диапазон цены.
func validatePriceRange(p float64) error {
	if p < AdPriceMin {
		return fmt.Errorf("цена не может быть меньше %.2f", AdPriceMin)
	}
	if p > AdPriceMax {
		return fmt.Errorf("цена не может превышать %.2f", AdPriceMax)
	}
	return nil
}

// validateImageURLBasic проверяет базовую корректность URL и допустимое расширение.
func validateImageURLBasic(raw string) error {
	if raw == "" {
		return errors.New("url не указан")
	}
	if len(raw) > AdImageURLMaxLen {
		return fmt.Errorf("длина url > %d", AdImageURLMaxLen)
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("некорректный url: %w", err)
	}
	ext := strings.ToLower(path.Ext(u.Path))
	if _, ok := allowedImageExt[ext]; !ok {
		return fmt.Errorf("недопустимый формат изображения: %s", ext)
	}
	return nil
}

// ValidateRemoteImage проверяет доступность URL, размер в байтах и (ограниченно) пиксельные размеры.
// ctx следует ограничивать таймаутом в вызывающем коде.
func ValidateRemoteImage(ctx context.Context, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return fmt.Errorf("head request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("image head: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("image недоступно: status %d", resp.StatusCode)
	}

	if resp.ContentLength > 0 && resp.ContentLength > AdImageMaxBytes {
		return fmt.Errorf("размер изображения > %d байт", AdImageMaxBytes)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "image/") {
		return fmt.Errorf("ожидался Content-Type image/*, получили %s", ct)
	}

	if err := sniffImageDimensions(ctx, rawURL); err != nil {
		return err
	}
	return nil
}

func sniffImageDimensions(ctx context.Context, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	const sniffLimit = 512 << 10 // 512KB
	lr := io.LimitReader(resp.Body, sniffLimit)

	cfg, _, err := image.DecodeConfig(lr)
	if err != nil {
		return fmt.Errorf("не удалось прочитать изображение: %w", err)
	}
	if cfg.Width > AdImageMaxWidth || cfg.Height > AdImageMaxHeight {
		return fmt.Errorf("изображение слишком большое (%dx%d), максимум %dx%d", cfg.Width, cfg.Height, AdImageMaxWidth, AdImageMaxHeight)
	}
	return nil
}
