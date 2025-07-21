package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/utils"
	"github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	DB *sql.DB
}

type ScanUser struct {
	ID           int
	Login        string
	Name         string
	Surname      string
	PasswordHash []byte
	PasswordSalt []byte
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
}

func (u *ScanUser) GetEntity() *entity.User {
	return &entity.User{
		ID:           u.ID,
		Login:        u.Login,
		Name:         u.Name,
		Surname:      u.Surname,
		PasswordHash: u.PasswordHash,
		PasswordSalt: u.PasswordSalt,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}

func NewUserRepository(db *sql.DB) (repository.UserRepository, error) {
	return &UserRepository{DB: db}, nil
}

func (r *UserRepository) Create(ctx context.Context, login, name, surname string, passwordHash, passwordSalt []byte) (*entity.User, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: создание пользователя")

	query := `
		INSERT INTO uuser (login, password_hashed, password_salt, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, login, password_hashed, password_salt, first_name, last_name
	`

	var createdUser entity.User
	err := r.DB.QueryRowContext(ctx, query,
		login,
		passwordHash,
		passwordSalt,
		name,
		surname,
	).Scan(
		&createdUser.ID,
		&createdUser.Login,
		&createdUser.PasswordHash,
		&createdUser.PasswordSalt,
		&createdUser.Name,
		&createdUser.Surname,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation
				return nil, fmt.Errorf("пользователь с таким логином уже существует: %w", err)
			case "23502": // not_null_violation
				return nil, fmt.Errorf("обязательное поле отсутствует: %w", err)
			case "23514": // check_violation
				return nil, fmt.Errorf("нарушено условие проверки: %w", err)
			}
		}

		logger.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при создании пользователя")

		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return &createdUser, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: получение пользователя по ID")

	query := `
		SELECT id, login, first_name, last_name, password_hashed, password_salt, created_at, updated_at
		FROM uuser
		WHERE id = $1
	`

	scanUser := ScanUser{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&scanUser.ID,
		&scanUser.Login,
		&scanUser.Name,
		&scanUser.Surname,
		&scanUser.PasswordHash,
		&scanUser.PasswordSalt,
		&scanUser.CreatedAt,
		&scanUser.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с id=%d не найден: %w", id, err)
		}

		logger.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"userID":    id,
			"error":     err,
		}).Error("Ошибка при получении пользователя")

		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return scanUser.GetEntity(), nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	requestID := utils.GetRequestID(ctx)

	logger.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("SQL запрос: получение пользователя по логину")

	query := `
		SELECT id, login, first_name, last_name, password_hashed, password_salt, created_at, updated_at
		FROM uuser
		WHERE login = $1
	`

	var scanUser ScanUser
	err := r.DB.QueryRowContext(ctx, query, login).Scan(
		&scanUser.ID,
		&scanUser.Login,
		&scanUser.Name,
		&scanUser.Surname,
		&scanUser.PasswordHash,
		&scanUser.PasswordSalt,
		&scanUser.CreatedAt,
		&scanUser.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с логином=%s не найден: %w", login, err)
		}

		logger.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"login":     login,
			"error":     err,
		}).Error("Ошибка при получении пользователя")

		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return scanUser.GetEntity(), nil
}
