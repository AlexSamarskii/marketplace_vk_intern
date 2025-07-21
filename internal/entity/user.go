package entity

import "time"

type User struct {
	ID           int       `db:"id"`
	Login        string    `db:"login" valid:"required,alphanum,length(3|50)"`
	Name         string    `db:"first_name" valid:"required,utfletter,length(1|100)"`
	Surname      string    `db:"last_name" valid:"required,utfletter,length(1|100)"`
	PasswordHash []byte    `db:"-" valid:"-"`
	PasswordSalt []byte    `db:"-" valid:"-"`
	CreatedAt    time.Time `db:"created_at" valid:"-"`
	UpdatedAt    time.Time `db:"updated_at" valid:"-"`
}
