package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/AlexSamarskii/marketplace_vk_intern/internal/entity"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/repository"
	"github.com/AlexSamarskii/marketplace_vk_intern/internal/utils"
	l "github.com/AlexSamarskii/marketplace_vk_intern/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	userSessionsPrefix = "user_sessions:"
)

type SessionRepository struct {
	conn             redis.Conn
	sessionAliveTime int
	ctx              context.Context
}

func NewSessionRepository(conn redis.Conn, ttl int) (repository.SessionRepository, error) {
	return &SessionRepository{
		conn:             conn,
		sessionAliveTime: ttl,
		ctx:              context.Background(),
	}, nil
}

func (r *SessionRepository) CreateSession(ctx context.Context, userID int) (string, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        userID,
	}).Info("создание сессии в Redis CreateSession")

	sessionToken := uuid.NewString()

	for {
		exists, err := redis.Int(r.conn.Do("EXISTS", sessionToken))
		if err != nil {
			return "", entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить сессию для пользователя с id=%d :%w", userID, err),
			)
		}
		if exists == 0 {
			break
		}
		sessionToken = uuid.NewString()
	}

	_, err := r.conn.Do("SET", sessionToken, fmt.Sprintf("%d", userID), "EX", r.sessionAliveTime)
	if err != nil {
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось создать сессию для пользователя с id=%d:%w", userID, err),
		)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID)
	_, err = r.conn.Do("SADD", userSessionsKey, sessionToken)
	if err != nil {
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось добавить сессию пользователя с id=%d в его активные сессии :%w", userID, err),
		)
	}

	_, err = r.conn.Do("EXPIRE", userSessionsKey, r.sessionAliveTime)
	if err != nil {
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить TTL на сессию пользователя с id=%d :%w", userID, err),
		)
	}

	return sessionToken, nil
}

func (r *SessionRepository) GetSession(ctx context.Context, sessionToken string) (int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":    requestID,
		"sessionToken": sessionToken,
	}).Info("получение сессии в Redis GetSession")

	reply, err := redis.String(r.conn.Do("GET", sessionToken))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return 0, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("не удалось найти сессию с токеном=%s :%w", sessionToken, err),
			)
		}
		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить сессию с токеном=%s :%w", sessionToken, err),
		)
	}

	var userID int
	_, err = fmt.Sscanf(reply, "%d", &userID)
	if err != nil {
		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось распарсить сессию на id с ключом=%s :%w", reply, err),
		)
	}

	return userID, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":    requestID,
		"sessionToken": sessionToken,
	}).Info("удаление сессии в Redis DeleteSession")

	reply, err := redis.String(r.conn.Do("GET", sessionToken))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil
		}
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить сессию с токеном=%s для удаления :%w", sessionToken, err),
		)
	}

	_, err = r.conn.Do("DEL", sessionToken)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить сессию с токеном=%s :%w", sessionToken, err),
		)
	}

	var userID int
	_, err = fmt.Sscanf(reply, "%d", &userID)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось распарсить сессию на id с ключом=%s :%w", reply, err),
		)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID)
	_, err = r.conn.Do("SREM", userSessionsKey, sessionToken)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить сессию с ключом=%s и токеном=%s из активных сессий пользователя :%w", userSessionsKey, sessionToken, err),
		)
	}

	return nil
}

func (r *SessionRepository) DeleteAllSessions(ctx context.Context, userID int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        userID,
	}).Info("удаление всех активных сессий пользователя в Redis DeleteAllSessions")

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID)

	sessions, err := redis.Strings(r.conn.Do("SMEMBERS", userSessionsKey))
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить активные сессии пользователя по ключу=%s :%w", userSessionsKey, err),
		)
	}

	for _, session := range sessions {
		_, err = r.conn.Do("DEL", session)
		if err != nil {
			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось удалить сессию из активные сессии пользователя c ключом=%s :%w", session, err),
			)
		}
	}

	_, err = r.conn.Do("DEL", userSessionsKey)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить ключ списка активных сессий пользователя c ключом=%s :%w", userSessionsKey, err),
		)
	}

	return nil
}
