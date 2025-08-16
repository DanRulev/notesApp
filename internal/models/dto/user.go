package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserCreate struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username" validate:"required,min=3,max=255"`
	Email    string    `json:"email" validate:"required,max=255,email"`
	Password string    `json:"password" validate:"required,min=8,max=72"`
	ImageURL string    `json:"image_url" validate:"omitempty,url"`
}

type UserUpdate struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	Username *string   `json:"username" validate:"omitempty,min=3,max=255"`
	Email    *string   `json:"email" validate:"omitempty,email,max=255"`
	ImageURL *string   `json:"image_url" validate:"omitempty,url"`
}

type UserSignIn struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type UserOutput struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserUpdPassword struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	OldPassword string    `json:"old_password" validate:"required,min=8,max=72"`
	NewPassword string    `json:"new_password" validate:"required,min=8,max=72"`
}
