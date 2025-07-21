package dto

import "time"

type UserProfileResponse struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"first_name"`
	Surname   string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
