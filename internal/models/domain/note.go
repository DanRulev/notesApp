package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Heading   string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteUpdate struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Heading *string
	Content *string
}

func (n *Note) Validate() error {
	if n.ID == uuid.Nil {
		return fmt.Errorf("invalid note ID")
	}

	if n.UserID == uuid.Nil {
		return fmt.Errorf("invalid note user ID")
	}

	if n.Heading == "" {
		return fmt.Errorf("empty heading")
	}

	if n.Content == "" {
		return fmt.Errorf("empty content")
	}

	return nil
}

func (n *NoteUpdate) Validate() error {
	if n.ID == uuid.Nil {
		return fmt.Errorf("invalid note ID")
	}

	if n.UserID == uuid.Nil {
		return fmt.Errorf("invalid note user ID")
	}

	if *n.Heading == "" {
		return fmt.Errorf("empty heading")
	}

	if *n.Content == "" {
		return fmt.Errorf("empty content")
	}

	return nil
}
