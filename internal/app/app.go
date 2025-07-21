package app

import (
	"net/http"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/config"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository/postgres"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository/redis"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/server"
	handler "github.com/AlexSamarskii/marketplace_vk_intern/internal/transport/http"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase/service"
	"github.com/AlexSamarskii/marketplace_vk_intern/pkg/connector"
	l "github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
)

func Init(cfg *config.Config) *server.Server {
	// Postgres Connection
	adConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to advertisement postgres: %v", err)
	}

	userConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to user postgres: %v", err)
	}

	// Redis Connection
	sessionConn, err := connector.NewRedisConnection(cfg.Redis)
	if err != nil {
		l.Log.Errorf("Failed to connect to session redis: %v", err)
	}

	// Repositories Init
	adRepo, err := postgres.NewAdvertisementRepository(adConn)
	if err != nil {
		l.Log.Errorf("Failed to create advertisement repository: %v", err)
	}

	userRepo, err := postgres.NewUserRepository(userConn)
	if err != nil {
		l.Log.Errorf("Failed to create user repository: %v", err)
	}

	sessionRepo, err := redis.NewSessionRepository(sessionConn, cfg.Redis.TTL)
	if err != nil {
		l.Log.Errorf("Failed to create session repository: %v", err)
	}

	// Use Cases Init
	authService := service.NewAuthService(sessionRepo, userRepo)
	userService := service.NewUserService(userRepo)
	adService := service.NewAdvertisementService(adRepo, userRepo)
	// Transport Init
	authHandler := handler.NewAuthHandler(authService, cfg.CSRF)
	userHandler := handler.NewUserHandler(authService, userService, cfg.CSRF)
	adHandler := handler.NewAdvertisementHandler(authService, adService, cfg.CSRF)

	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		userHandler.Configure(r)
		adHandler.Configure(r)
	})

	return srv
}
