package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var errorToStatus = map[error]int{
	entity.ErrNotFound:      http.StatusNotFound,
	entity.ErrBadRequest:    http.StatusBadRequest,
	entity.ErrUnauthorized:  http.StatusUnauthorized,
	entity.ErrForbidden:     http.StatusForbidden,
	entity.ErrAlreadyExists: http.StatusConflict,
	entity.ErrInternal:      http.StatusInternalServerError,
}

func ToAPIError(err error) APIError {
	var apiError APIError
	var customError *entity.Error

	if errors.As(err, &customError) {
		client := customError.ClientErr()
		apiError.Message = customError.ClientErr().Error()

		if status, ok := errorToStatus[client]; ok {
			apiError.Status = status
		} else {
			apiError.Status = http.StatusInternalServerError
		}
		return apiError
	}

	apiError.Message = "internal server error"
	apiError.Status = http.StatusInternalServerError
	return apiError
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteAPIError(w, APIError{
		Status:  status,
		Message: err.Error(),
	})
}

func WriteAPIError(w http.ResponseWriter, apiError APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Status)
	err := json.NewEncoder(w).Encode(apiError)
	if err != nil {
		return
	}
}
