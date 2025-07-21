package http

import (
	"encoding/json"
	"net/http"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/config"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/middleware"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/transport/http/utils"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase"
)

type AuthHandler struct {
	auth usecase.AuthUsecase
	cfg  config.CSRFConfig
}

func NewAuthHandler(auth usecase.AuthUsecase, cfg config.CSRFConfig) AuthHandler {
	return AuthHandler{auth: auth, cfg: cfg}
}

func (h *AuthHandler) Configure(r *http.ServeMux) {
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /isAuth", h.IsAuth)
	authMux.HandleFunc("POST /logout", h.Logout)
	authMux.HandleFunc("POST /logoutAll", h.LogoutAll)

	r.Handle("/auth/", http.StripPrefix("/auth", authMux))
}

// IsAuth godoc
// @Tags Auth
// @Summary Проверка авторизации
// @Description Проверяет авторизован пользователь или нет.
// @Security session_cookie
// @Produce json
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /auth/isAuth [get]
func (h *AuthHandler) IsAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(dto.AuthResponse{UserId: userID}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// Logout godoc
// @Tags Auth
// @Summary Выход из системы
// @Description Завершает текущую сессию пользователя
// @Success 200
// @Failure 500 {object} utils.APIError
// @Router /auth/logout [post]
// @Security session_cookie
// @Security csrf_token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.auth.Logout(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// очищаем старые cookie
	utils.ClearTokenCookies(w)
	// устанавливаем новый токен
	middleware.SetCSRFToken(w, r, h.cfg)
	w.WriteHeader(http.StatusOK)
}

// LogoutAll godoc
// @Tags Auth
// @Summary Выход со всех устройств
// @Description Завершает все активные сессии пользователя
// @Success 200
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /auth/logoutAll [post]
// @Security session_cookie
// @Security csrf_token
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := h.auth.LogoutAll(ctx, userID); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// очищаем старые cookie
	utils.ClearTokenCookies(w)
	// устанавливаем новый токен
	middleware.SetCSRFToken(w, r, h.cfg)
	w.WriteHeader(http.StatusOK)
}
