package middleware

import (
	"net/http"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/utils"
	"github.com/google/uuid"
)

func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.NewString()
			ctx := utils.SetRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
