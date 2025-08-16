package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Password  string
	ImageURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserUpdate struct {
	ID       uuid.UUID
	Username *string
	Email    *string
	Password *string
	ImageURL *string
}

func (u User) Validate() error {
	if u.ID == uuid.Nil {
		return fmt.Errorf("invalid user ID")
	}

	if u.Username == "" {
		return fmt.Errorf("empty username")
	}

	if u.Email == "" {
		return fmt.Errorf("empty email")
	}

	if u.Password == "" {
		return fmt.Errorf("empty password")
	}

	return nil
}

func (uu *UserUpdate) Validate() error {
	if uu.ID == uuid.Nil {
		return fmt.Errorf("invalid user ID")
	}

	if uu.Username != nil && *uu.Username == "" {
		return fmt.Errorf("empty username")
	}

	if uu.Email != nil && *uu.Email == "" {
		return fmt.Errorf("empty email")
	}

	if uu.Password != nil && *uu.Password == "" {
		return fmt.Errorf("empty password")
	}

	return nil
}
