package usecase

import (
	"context"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
)

type UserUsecase interface {
	Register(ctx context.Context, registerDTO *dto.UserRegister) (*dto.UserProfileResponse, error)
	Login(ctx context.Context, loginDTO *dto.Login) (int, error)
	GetUser(ctx context.Context, employerID int) (*dto.UserProfileResponse, error)
	LoginExists(ctx context.Context, email string) (*dto.LoginExistsResponse, error)
}
