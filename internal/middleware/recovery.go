package middleware

import (
	"fmt"
	"net/http"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/transport/http/utils"
	GlobalUtils "github.com/AlexSamarskii/marketplace_vk_intern/internal/utils"
	l "github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
	"github.com/sirupsen/logrus"
)

func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("%v", err))
					requestID := GlobalUtils.GetRequestID(r.Context())
					l.Log.WithFields(logrus.Fields{
						"requestID": requestID,
					}).Fatal("Recovery middleware panic")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
