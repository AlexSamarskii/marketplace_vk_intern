package entity

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"regexp"

	"golang.org/x/crypto/argon2"
)

const (
	TimeCost        = 2
	MemoryCost      = 64 * 1024
	ParallelThreads = 2
	HashLength      = 32
)

func ValidatePassword(password string) error {
	switch {
	case len(password) < 8:
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать не менее 8 символов"),
		)

	case len(password) > 32:
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать не более 32 символов"),
		)

	case !regexp.MustCompile(`^[!@#$%^&*a-zA-Z0-9_]+$`).MatchString(password):
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен состоять из латинских букв, цифр и специальных символов !@#$%%^&*"),
		)

	default:
		return nil
	}
}

func HashPassword(password string) (salt []byte, hash []byte, err error) {
	salt = make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, NewError(
			ErrInternal,
			fmt.Errorf("ошибка при хешировании пароля"),
		)
	}

	hash = argon2.IDKey(
		[]byte(password),
		salt,
		TimeCost,
		MemoryCost,
		ParallelThreads,
		HashLength,
	)
	return salt, hash, nil
}

func CheckPassword(password string, passwordHash, passwordSalt []byte) bool {
	return bytes.Equal(
		argon2.IDKey(
			[]byte(password),
			passwordSalt,
			TimeCost,
			MemoryCost,
			ParallelThreads,
			HashLength,
		),
		passwordHash,
	)
}

func ValidateLogin(login string) error {
	re := regexp.MustCompile(`^[A-Za-z0-9._-]{3,30}$`)

	if !re.MatchString(login) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("невалидный логин"),
		)
	}

	if len(login) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("логин не может быть длиннее 255 символов"),
		)
	}
	return nil
}
