package service

import (
	"context"
	"errors"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase"
)

type AuthService struct {
	sessionRepository repository.SessionRepository
	userRepository    repository.UserRepository
}

func NewAuthService(
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
) usecase.AuthUsecase {
	return &AuthService{
		sessionRepository: sessionRepo,
		userRepository:    userRepo,
	}
}

func (a *AuthService) EmailExists(ctx context.Context, login string) (*dto.LoginExistsResponse, error) {
	if err := entity.ValidateLogin(login); err != nil {
		return nil, err
	}

	applicant, err := a.userRepository.GetByLogin(ctx, login)
	if err == nil && applicant != nil {
		return &dto.LoginExistsResponse{
			Exists: true,
		}, err
	}

	var e entity.Error
	if errors.As(err, &e) && !errors.Is(e.ClientErr(), entity.ErrNotFound) {
		return nil, err
	}

	employer, err := a.userRepository.GetByLogin(ctx, login)
	if err == nil && employer != nil {
		return &dto.LoginExistsResponse{
			Exists: true,
		}, err
	}

	return nil, err
}

func (a *AuthService) Logout(ctx context.Context, session string) error {
	if err := a.sessionRepository.DeleteSession(ctx, session); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) LogoutAll(ctx context.Context, userID int) error {
	if err := a.sessionRepository.DeleteAllSessions(ctx, userID); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) GetUserIDBySession(ctx context.Context, session string) (int, error) {
	userID, err := a.sessionRepository.GetSession(ctx, session)
	if err != nil {
		return -1, err
	}
	return userID, nil
}

func (a *AuthService) CreateSession(ctx context.Context, userID int) (string, error) {
	session, err := a.sessionRepository.CreateSession(ctx, userID)
	if err != nil {
		return "", err
	}
	return session, nil
}
