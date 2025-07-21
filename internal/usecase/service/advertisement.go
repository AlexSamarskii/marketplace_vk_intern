package service

import (
	"context"
	"fmt"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/utils"
	"github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
	"github.com/AlexSamarskii/marketplace_vk_intern/pkg/sanitizer"
	"github.com/sirupsen/logrus"
)

type AdvertisementService struct {
	adRepo   repository.AdvertisementRepository
	userRepo repository.UserRepository
}

func NewAdvertisementService(
	adRepo repository.AdvertisementRepository,
	userRepo repository.UserRepository,
) usecase.AdvertisementUsecase {
	return &AdvertisementService{
		adRepo:   adRepo,
		userRepo: userRepo,
	}
}

func (s *AdvertisementService) Create(ctx context.Context, userID int, req *dto.CreateAdvertisementRequest) (*dto.AdvertisementShort, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
	}).Info("Создание объявления")

	req.Title = sanitizer.StrictPolicy.Sanitize(req.Title)
	req.Description = sanitizer.StrictPolicy.Sanitize(req.Description)
	req.ImageURL = sanitizer.StrictPolicy.Sanitize(req.ImageURL)

	ad := &entity.Advertisement{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
	}

	if _, err := ad.Validate(); err != nil {
		return nil, entity.NewError(entity.ErrBadRequest,
			fmt.Errorf("ошибка валидации объявления: %w", err))
	}

	createdAd, err := s.adRepo.Create(ctx, ad)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при создании объявления")
		return nil, err
	}

	response := &dto.AdvertisementShort{
		Title:       createdAd.Title,
		Description: createdAd.Description,
		ImageURL:    createdAd.ImageURL,
		Price:       createdAd.Price,
		CreatedAt:   createdAd.CreatedAt,
		UpdatedAt:   createdAd.UpdatedAt,
	}

	return response, nil
}

func (s *AdvertisementService) GetByID(ctx context.Context, id int) (*dto.AdvertisementShort, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"adID":      id,
	}).Info("Получение объявления по ID")

	ad, err := s.adRepo.GetByID(ctx, id)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"adID":      id,
			"error":     err,
		}).Error("Ошибка при получении объявления")
		return nil, fmt.Errorf("ошибка при получении объявления: %w", err)
	}

	response := &dto.AdvertisementShort{
		Title:       ad.Title,
		Description: ad.Description,
		ImageURL:    ad.ImageURL,
		Price:       ad.Price,
		CreatedAt:   ad.CreatedAt,
		UpdatedAt:   ad.UpdatedAt,
	}

	return response, nil
}

func (s *AdvertisementService) GetAll(
	ctx context.Context,
	userID int,
	offset, limit int,
	sortBy, order string,
	minPrice, maxPrice *float64,
) ([]dto.AdvertisementResponse, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Получение списка объявлений")

	ads, err := s.adRepo.GetAll(ctx, userID, offset, limit, sortBy, order, minPrice, maxPrice)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка объявлений: %w", err)
	}

	response := make([]dto.AdvertisementResponse, 0, len(ads))
	for _, ad := range ads {
		response = append(response, dto.AdvertisementResponse{
			ID:          ad.ID,
			Title:       ad.Title,
			Description: ad.Description,
			ImageURL:    ad.ImageURL,
			Price:       ad.Price,
			AuthorLogin: ad.AuthorLogin,
			IsMine:      ad.IsMine && userID != 0,
			CreatedAt:   ad.CreatedAt,
			UpdatedAt:   ad.UpdatedAt,
		})
	}

	return response, nil
}

func (s *AdvertisementService) GetByUserID(ctx context.Context, userID int) ([]dto.AdvertisementResponse, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
	}).Info("Получение объявлений пользователя")

	ads, err := s.adRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при получении объявлений пользователя")
		return nil, fmt.Errorf("ошибка при получении объявлений пользователя: %w", err)
	}

	response := make([]dto.AdvertisementResponse, 0, len(ads))
	for _, ad := range ads {
		response = append(response, dto.AdvertisementResponse{
			ID:          ad.ID,
			Title:       ad.Title,
			Description: ad.Description,
			ImageURL:    ad.ImageURL,
			Price:       ad.Price,
			UserID:      ad.UserID,
			CreatedAt:   ad.CreatedAt,
			UpdatedAt:   ad.UpdatedAt,
		})
	}

	return response, nil
}
