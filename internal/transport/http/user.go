package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/config"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/middleware"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/transport/http/utils"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase"
)

type UserHandler struct {
	auth usecase.AuthUsecase
	user usecase.UserUsecase
	cfg  config.CSRFConfig
}

func NewUserHandler(auth usecase.AuthUsecase, user usecase.UserUsecase, cfg config.CSRFConfig) UserHandler {
	return UserHandler{auth: auth, user: user, cfg: cfg}
}

func (h *UserHandler) Configure(r *http.ServeMux) {
	userMux := http.NewServeMux()

	userMux.HandleFunc("POST /register", h.Register)
	userMux.HandleFunc("POST /login", h.Login)
	userMux.HandleFunc("GET /profile/{id}", h.GetProfile)

	r.Handle("/user/", http.StripPrefix("/user", userMux))
}

// Register godoc
// @Tags User
// @Summary Регистрация пользователя
// @Accept json
// @Produce json
// @Param registerData body dto.UserRegister true "Данные для регистрации"
// @Header 200 {string} Set-Cookie "Сессионные cookies"
// @Header 200 {string} X-CSRF-Token "CSRF-токен"
// @Success 200 {object} dto.UserProfileResponse
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 409 {object} utils.APIError "Пользователь уже существует"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /user/register [post]
// @Security csrf_token
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var registerDTO dto.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&registerDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	user, err := h.user.Register(ctx, &registerDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.CreateSession(w, r, h.auth, user.ID); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	middleware.SetCSRFToken(w, r, h.cfg)

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(dto.UserProfileResponse{ID: user.ID,
		Login:     user.Login,
		Name:      user.Name,
		Surname:   user.Surname,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// Login godoc
// @Tags User
// @Summary Авторизация пользователя
// @Description Авторизация пользователя. При успешной авторизации отправляет куки с сессией.
// Если пользователь уже авторизован, предыдущие cookies с сессией перезаписываются.
// Также устанавливает CSRF-токен при успешной авторизации.
// @Accept json
// @Produce json
// @Param loginData body dto.Login true "Данные для авторизации (login и пароль)"
// @Header 200 {string} Set-Cookie "Сессионные cookies"
// @Header 200 {string} X-CSRF-Token "CSRF-токен"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 403 {object} utils.APIError "Доступ запрещен (неверные учетные данные)"
// @Failure 404 {object} utils.APIError "Пользователь не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /user/login [post]
// @Security csrf_token
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var loginDTO dto.Login
	if err := json.NewDecoder(r.Body).Decode(&loginDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	userID, err := h.user.Login(ctx, &loginDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.CreateSession(w, r, h.auth, userID); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	middleware.SetCSRFToken(w, r, h.cfg)

	csrfToken := w.Header().Get("X-CSRF-Token")

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(dto.LoginResponse{Token: csrfToken}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// GetProfile godoc
// @Tags User
// @Summary Получить профиль пользователя
// @Description Возвращает профиль пользователя по ID. Требует авторизации. Доступен только для владельца профиля.
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} dto.UserProfileResponse "Профиль пользователя"
// @Failure 400 {object} utils.APIError "Неверный ID"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Нет доступа к этому профилю"
// @Failure 404 {object} utils.APIError "Профиль не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /user/profile/{id} [get]
// @Security session_cookie
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	_, err = h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	requestedID := r.PathValue("id")
	applicantID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	user, err := h.user.GetUser(ctx, applicantID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(user); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
