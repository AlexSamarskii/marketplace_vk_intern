package usecase

import (
	"context"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
)

type AuthUsecase interface {
	Logout(context.Context, string) error
	LogoutAll(context.Context, int) error
	GetUserIDBySession(context.Context, string) (int, error)
	CreateSession(context.Context, int) (string, error)
	EmailExists(context.Context, string) (*dto.LoginExistsResponse, error)
}
