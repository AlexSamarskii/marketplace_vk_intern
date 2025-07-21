package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/app"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/config"
	l "github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
)

// @title Marketplace API
// @version 1.0.0
// @description API веб-приложения Marketplace для публикации объявлений.
// @BasePath  /api/v1
// @securityDefinitions.apikey csrf_token
// @in header
// @name X-CSRF-Token
// @securityDefinitions.apikey session_cookie
// @in cookie
// @name session_id
func main() {

	cfg, err := config.Load()
	if err != nil {
		l.Log.Fatalf("Failed to load config: %v", err)
	}

	srv := app.Init(cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		l.Log.Info("Shutting down server...")
		if err := srv.Stop(); err != nil {
			l.Log.Fatalf("Failed to stop server: %v", err)
		}
	}()

	l.Log.Infof("Starting server on %s", cfg.HTTP.Port)
	l.Log.Infof("Server host: %s", cfg.HTTP.Host)
	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Log.Fatalf("Failed to run server: %v", err)
	}
}
