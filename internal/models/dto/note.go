package dto

import (
	"time"

	"github.com/google/uuid"
)

type NoteCreate struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Heading string    `json:"heading" validate:"required,min=1,max=255"`
	Content string    `json:"content" validate:"required,min=1,max=255"`
	Done    bool      `json:"done"`
}

type NoteUpdate struct {
	ID      uuid.UUID `json:"id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Heading *string   `json:"heading" validate:"omitempty,min=1,max=255"`
	Content *string   `json:"content" validate:"omitempty,min=1,max=255"`
	Done    *bool     `json:"done"`
}

type NoteOutput struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Heading   string    `json:"heading"`
	Content   string    `json:"content"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
