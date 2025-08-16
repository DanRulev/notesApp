package domain

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Token struct {
	UserID    uuid.UUID
	TokenID   string
	ExpiresAt time.Time
}

type TokenClaims struct {
	UserID uuid.UUID
	jwt.StandardClaims
}

func (t Token) Validate() error {
	if t.UserID == uuid.Nil {
		return fmt.Errorf("invalid token user ID")
	}

	if t.TokenID == uuid.Nil.String() {
		return fmt.Errorf("invalid token ID")
	}

	if t.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("invalid expires")
	}

	return nil
}
