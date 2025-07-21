package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/utils"
	l "github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type AdvertisementRepository struct {
	DB *sql.DB
}

func NewAdvertisementRepository(db *sql.DB) (repository.AdvertisementRepository, error) {
	return &AdvertisementRepository{DB: db}, nil
}

func (r *AdvertisementRepository) Create(ctx context.Context, ad *entity.Advertisement) (*entity.Advertisement, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: создание объявления")

	query := `
		INSERT INTO advertisement (
			user_id, title, description, image_url, price, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, user_id, title, description, image_url, price, created_at, updated_at
	`

	var createdAd entity.Advertisement
	err := r.DB.QueryRowContext(
		ctx,
		query,
		ad.UserID,
		ad.Title,
		ad.Description,
		ad.ImageURL,
		ad.Price,
	).Scan(
		&createdAd.ID,
		&createdAd.UserID,
		&createdAd.Title,
		&createdAd.Description,
		&createdAd.ImageURL,
		&createdAd.Price,
		&createdAd.CreatedAt,
		&createdAd.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation
				return nil, entity.NewError(entity.ErrAlreadyExists,
					fmt.Errorf("объявление с такими параметрами уже существует: %w", err))
			case "23502": // not_null_violation
				return nil, entity.NewError(entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует: %w", err))
			case "23514": // check_violation
				return nil, entity.NewError(entity.ErrBadRequest,
					fmt.Errorf("нарушено условие проверки: %w", err))
			case "23503": // foreign_key_violation
				return nil, entity.NewError(entity.ErrBadRequest,
					fmt.Errorf("указан несуществующий пользователь: %w", err))
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при создании объявления")

		return nil, entity.NewError(entity.ErrInternal,
			fmt.Errorf("ошибка при создании объявления: %w", err))
	}

	return &createdAd, nil
}

func (r *AdvertisementRepository) GetByID(ctx context.Context, id int) (*entity.Advertisement, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: получение объявления по ID")

	query := `
		SELECT 
			a.id, a.user_id, a.title, a.description, a.image_url, a.price, 
			a.created_at, a.updated_at
		FROM advertisement a
		WHERE a.id = $1
	`

	var ad entity.Advertisement
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&ad.ID,
		&ad.UserID,
		&ad.Title,
		&ad.Description,
		&ad.ImageURL,
		&ad.Price,
		&ad.CreatedAt,
		&ad.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("объявление с id=%d не найдено: %w", id, err)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"adID":      id,
			"error":     err,
		}).Error("Ошибка при получении объявления")

		return nil, fmt.Errorf("ошибка при получении объявления: %w", err)
	}

	return &ad, nil
}

func (r *AdvertisementRepository) GetAll(ctx context.Context, userID int, offset, limit int, sortBy, order string, minPrice, maxPrice *float64) ([]entity.Advertisement, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: получение списка объявлений")

	if sortBy != "created_at" && sortBy != "price" {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	whereParts := []string{"1=1"}
	args := []interface{}{userID} // userID = $1
	argPos := 2

	if minPrice != nil {
		whereParts = append(whereParts, fmt.Sprintf("a.price >= $%d", argPos))
		args = append(args, *minPrice)
		argPos++
	}
	if maxPrice != nil {
		whereParts = append(whereParts, fmt.Sprintf("a.price <= $%d", argPos))
		args = append(args, *maxPrice)
		argPos++
	}

	whereClause := strings.Join(whereParts, " AND ")

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)

	q := fmt.Sprintf(`
        SELECT 
            a.id, a.user_id, a.title, a.description, a.image_url, a.price,
            a.created_at, a.updated_at, u.login AS author_login,
            (a.user_id = $1) AS is_mine
        FROM advertisement a
        JOIN uuser u ON a.user_id = u.id
        WHERE %s
        ORDER BY a.%s %s
        LIMIT $%d OFFSET $%d
    `, whereClause, sortBy, order, limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка объявлений: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var ads []entity.Advertisement
	for rows.Next() {
		var ad entity.Advertisement
		err := rows.Scan(
			&ad.ID,
			&ad.UserID,
			&ad.Title,
			&ad.Description,
			&ad.ImageURL,
			&ad.Price,
			&ad.CreatedAt,
			&ad.UpdatedAt,
			&ad.AuthorLogin,
			&ad.IsMine,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("Ошибка при сканировании объявления")

			return nil, fmt.Errorf("ошибка при сканировании объявления: %w", err)
		}
		ads = append(ads, ad)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при итерации по объявлениям")

		return nil, fmt.Errorf("ошибка при итерации по объявлениям: %w", err)
	}

	return ads, nil
}

func (r *AdvertisementRepository) GetByUserID(ctx context.Context, userID int) ([]entity.Advertisement, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: получение объявлений пользователя")

	query := `
		SELECT 
			id, user_id, title, description, image_url, price, 
			created_at, updated_at
		FROM advertisement
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"userID":    userID,
			"error":     err,
		}).Error("Ошибка при получении объявлений пользователя")

		return nil, fmt.Errorf("ошибка при получении объявлений пользователя: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var ads []entity.Advertisement
	for rows.Next() {
		var ad entity.Advertisement
		err := rows.Scan(
			&ad.ID,
			&ad.UserID,
			&ad.Title,
			&ad.Description,
			&ad.ImageURL,
			&ad.Price,
			&ad.CreatedAt,
			&ad.UpdatedAt,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("Ошибка при сканировании объявления")

			return nil, fmt.Errorf("ошибка при сканировании объявления: %w", err)
		}
		ads = append(ads, ad)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при итерации по объявлениям")

		return nil, fmt.Errorf("ошибка при итерации по объявлениям: %w", err)
	}

	return ads, nil
}
