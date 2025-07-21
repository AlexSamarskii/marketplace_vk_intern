package repository

import (
	"context"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
)

type AdvertisementRepository interface {
	Create(ctx context.Context, ad *entity.Advertisement) (*entity.Advertisement, error)
	GetByID(ctx context.Context, id int) (*entity.Advertisement, error)
	GetAll(ctx context.Context, userID int, page, limit int, sortBy, order string, minPrice, maxPrice *float64) ([]entity.Advertisement, error)
	GetByUserID(ctx context.Context, userID int) ([]entity.Advertisement, error)
}
