package entity

import "time"

type User struct {
	ID           int       `db:"id"`
	Login        string    `db:"login"`
	Name         string    `db:"first_name"`
	Surname      string    `db:"last_name"`
	PasswordHash []byte    `db:"-"`
	PasswordSalt []byte    `db:"-"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
