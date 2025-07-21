package usecase

import (
	"context"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
)

type AdvertisementUsecase interface {
	Create(ctx context.Context, userID int, req *dto.CreateAdvertisementRequest) (*dto.AdvertisementShort, error)
	GetByID(ctx context.Context, id int) (*dto.AdvertisementShort, error)
	GetAll(ctx context.Context, userID int, page, limit int, sortBy, order string, minPrice, maxPrice *float64) ([]dto.AdvertisementResponse, error)
	GetByUserID(ctx context.Context, userID int) ([]dto.AdvertisementResponse, error)
}
