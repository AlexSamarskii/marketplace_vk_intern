package repository

import "context"

type SessionRepository interface {
	CreateSession(ctx context.Context, userID int) (string, error)
	GetSession(ctx context.Context, sessionToken string) (userID int, err error)
	DeleteSession(ctx context.Context, sessionToken string) error
	DeleteAllSessions(ctx context.Context, userID int) error
}
