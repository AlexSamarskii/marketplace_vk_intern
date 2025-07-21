package service

import (
	"context"
	"fmt"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity/dto"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/usecase"
	"github.com/AlexSamarskii/marketplace_vk_intern/pkg/sanitizer"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(
	userRepo repository.UserRepository,
) usecase.UserUsecase {
	return &UserService{
		userRepo: userRepo,
	}
}

func (e *UserService) Register(ctx context.Context, registerDTO *dto.UserRegister) (*dto.UserProfileResponse, error) {
	user := new(entity.User)

	if err := entity.ValidateLogin(registerDTO.Login); err != nil {
		return nil, err
	}

	if err := entity.ValidatePassword(registerDTO.Password); err != nil {
		return nil, err
	}

	salt, hash, err := entity.HashPassword(registerDTO.Password)
	if err != nil {
		return nil, err
	}

	sanitizedName := sanitizer.StrictPolicy.Sanitize(registerDTO.Name)
	sanitizedSurname := sanitizer.StrictPolicy.Sanitize(registerDTO.Surname)
	user, err = e.userRepo.Create(ctx, registerDTO.Login, sanitizedName, sanitizedSurname, hash, salt)
	if err != nil {
		return nil, err
	}

	response := &dto.UserProfileResponse{
		ID:        user.ID,
		Login:     user.Login,
		Name:      user.Name,
		Surname:   user.Surname,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	return response, nil
}

func (e *UserService) Login(ctx context.Context, loginDTO *dto.Login) (int, error) {
	if err := entity.ValidateLogin(loginDTO.Login); err != nil {
		return -1, err
	}

	if err := entity.ValidatePassword(loginDTO.Password); err != nil {
		return -1, err
	}

	employer, err := e.userRepo.GetByLogin(ctx, loginDTO.Login)
	if err != nil {
		return -1, err
	}
	if entity.CheckPassword(loginDTO.Password, employer.PasswordHash, employer.PasswordSalt) {
		return employer.ID, nil
	}
	return -1, entity.NewError(
		entity.ErrForbidden,
		fmt.Errorf("неверный пароль"),
	)
}

func (e *UserService) GetUser(ctx context.Context, employerID int) (*dto.UserProfileResponse, error) {
	employer, err := e.userRepo.GetByID(ctx, employerID)
	if err != nil {
		return nil, err
	}
	return e.employerEntityToDTO(ctx, employer)
}

func (e *UserService) employerEntityToDTO(ctx context.Context, employer *entity.User) (*dto.UserProfileResponse, error) {
	profile := &dto.UserProfileResponse{
		ID:        employer.ID,
		Name:      employer.Name,
		Surname:   employer.Surname,
		Login:     employer.Login,
		CreatedAt: employer.CreatedAt,
		UpdatedAt: employer.UpdatedAt,
	}
	return profile, nil
}

func (e *UserService) LoginExists(ctx context.Context, email string) (*dto.LoginExistsResponse, error) {
	if err := entity.ValidateLogin(email); err != nil {
		return nil, err
	}

	applicant, err := e.userRepo.GetByLogin(ctx, email)
	if err == nil && applicant != nil {
		return &dto.LoginExistsResponse{
			Exists: true,
		}, nil
	}

	return nil, err
}
