package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/config"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/transport/http/utils"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase"
	"github.com/AlexSamarskii/marketplace_vk_intern/pkg/sanitizer"
)

type AdvertisementHandler struct {
	auth          usecase.AuthUsecase
	advertisement usecase.AdvertisementUsecase
	cfg           config.CSRFConfig
}

func NewAdvertisementHandler(auth usecase.AuthUsecase, ad usecase.AdvertisementUsecase, cfg config.CSRFConfig) AdvertisementHandler {
	return AdvertisementHandler{auth: auth, advertisement: ad, cfg: cfg}
}

func (h *AdvertisementHandler) Configure(r *http.ServeMux) {
	adMux := http.NewServeMux()

	adMux.HandleFunc("POST /create", h.CreateAdvertisement)
	adMux.HandleFunc("GET /{id}", h.GetAdvertisement)
	adMux.HandleFunc("GET /all", h.GetAllAdvertisements)

	r.Handle("/ad/", http.StripPrefix("/ad", adMux))
}

// CreateAdvertisement godoc
// @Tags Advertisement
// @Summary Создание нового объявления
// @Description Создает новое объявления для авторизованного пользователя. Требует авторизации и CSRF-токена.
// @Accept json
// @Produce json
// @Param advertisementData body dto.CreateAdvertisementRequest true "Данные для создания объявления"
// @Success 201 {object} dto.AdvertisementShort "Созданное объявление"
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещен (только для соискателей)"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /ad/create [post]
// @Security csrf_token
// @Security session_cookie
func (h *AdvertisementHandler) CreateAdvertisement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	var createAdRequest dto.CreateAdvertisementRequest
	if err := json.NewDecoder(r.Body).Decode(&createAdRequest); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	createAdRequest.Title = sanitizer.StrictPolicy.Sanitize(createAdRequest.Title)
	createAdRequest.Description = sanitizer.StrictPolicy.Sanitize(createAdRequest.Description)
	createAdRequest.ImageURL = sanitizer.StrictPolicy.Sanitize(createAdRequest.ImageURL)

	ad, err := h.advertisement.Create(ctx, userID, &createAdRequest)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// GetAdvertisement godoc
// @Tags Advertisement
// @Summary Получение объявления по ID
// @Description Возвращает полную информацию об объявлении по его ID. Доступно всем авторизованным пользователям.
// @Produce json
// @Param id path int true "ID объявления"
// @Success 200 {object} dto.AdvertisementShort "Информация об объявлении"
// @Failure 400 {object} utils.APIError "Неверный ID"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 404 {object} utils.APIError "Объявление не найдено"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /ad/{id} [get]
// @Security session_cookie
func (h *AdvertisementHandler) GetAdvertisement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получение ID объявления из URL
	adIDStr := r.PathValue("id")
	adID, err := strconv.Atoi(adIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	ad, err := h.advertisement.GetByID(ctx, adID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// GetAllAdvertisements godoc
// @Tags Advertisement
// @Summary Получение всех объявлений
// @Description Возвращает список объявлений с поддержкой пагинации, сортировки и фильтрации по цене.
// @Produce json
// @Param limit query int false "Количество объявлений на странице (по умолчанию 10)"
// @Param offset query int false "Смещение от начала списка (по умолчанию 0)"
// @Param sort query string false "Поле сортировки (created_at или price)"
// @Param order query string false "Направление сортировки (asc или desc)"
// @Param min_price query number false "Минимальная цена фильтрации"
// @Param max_price query number false "Максимальная цена фильтрации"
// @Success 200 {object} []dto.AdvertisementResponse "Список объявлений"
// @Failure 400 {object} utils.APIError "Некорректные параметры запроса"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /ad/all [get]
func (h *AdvertisementHandler) GetAllAdvertisements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	var userID int
	if cookie, err := r.Cookie("session_id"); err == nil && cookie != nil {
		if uid, err := h.auth.GetUserIDBySession(ctx, cookie.Value); err == nil {
			userID = uid
		}
	}

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil || l <= 0 || l > 100 {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
		limit = l
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		o, err := strconv.Atoi(v)
		if err != nil || o < 0 {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
		offset = o
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "created_at"
	}
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	var (
		minPricePtr *float64
		maxPricePtr *float64
	)
	if v := r.URL.Query().Get("min_price"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
		minPricePtr = &f
	}
	if v := r.URL.Query().Get("max_price"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
		maxPricePtr = &f
	}

	ads, err := h.advertisement.GetAll(ctx, userID, offset, limit, sortBy, order, minPricePtr, maxPricePtr)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
