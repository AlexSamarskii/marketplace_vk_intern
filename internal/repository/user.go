package repository

import (
	"context"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, login, name, surname string, passwordHash, passwordSalt []byte) (*entity.User, error)
	GetByID(ctx context.Context, id int) (*entity.User, error)
	GetByLogin(ctx context.Context, login string) (*entity.User, error)
}
