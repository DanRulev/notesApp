package service

import (
	"context"
	"errors"
	"fmt"
	"noteApp/internal/config"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

type AuthRI interface {
	CreateUser(ctx context.Context, user domain.User) error
	CreateToken(ctx context.Context, token domain.Token) error
	UserCredentials(ctx context.Context, email string) (uuid.UUID, string, error)
	Token(ctx context.Context, tokenID string) (domain.Token, error)
	DeleteToken(ctx context.Context, tokenID string) error
}

type AuthS struct {
	repo   AuthRI
	token  config.AuthCfg
	hasher HasherI
	log    *logger.Logger
}

func NewAuthService(
	repo AuthRI,
	hasher HasherI,
	token config.AuthCfg,
	log *logger.Logger,
) *AuthS {
	return &AuthS{
		repo:   repo,
		hasher: hasher,
		token:  token,
		log:    log,
	}
}

func (a *AuthS) SignUp(ctx context.Context, user dto.UserCreate) (uuid.UUID, error) {
	hashPass, err := a.hasher.GenerateHash(user.Password)
	if err != nil {
		a.log.Error("failed to hash password during sign-up",
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	input := domain.User{
		ID:       uuid.New(),
		Username: user.Username,
		Email:    user.Email,
		Password: hashPass,
		ImageURL: user.ImageURL,
	}

	if err := input.Validate(); err != nil {
		a.log.Debug("sign-up validation failed",
			zap.String("email", user.Email),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	if err := a.repo.CreateUser(ctx, input); err != nil {
		a.log.Error("failed to create user in repository",
			zap.String("email", user.Email),
			zap.String("user_id", input.ID.String()),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	a.log.Info("user signed up successfully",
		zap.String("user_id", input.ID.String()),
		zap.String("email", user.Email),
	)

	return input.ID, nil
}

func (a *AuthS) SignIn(ctx context.Context, data dto.UserSignIn) (dto.TokenOutput, error) {
	userID, err := a.checkUser(ctx, data.Email, data.Password)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			a.log.Warn("invalid credentials during sign-in",
				zap.String("email", data.Email),
			)
		} else {
			a.log.Error("unexpected error during sign-in",
				zap.String("email", data.Email),
				zap.Error(err),
			)
		}
		return dto.TokenOutput{}, err
	}

	token, err := a.generateAndSaveTokens(ctx, userID)
	if err != nil {
		a.log.Error("failed to generate or save tokens",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return dto.TokenOutput{}, err
	}

	return token, nil
}

func (a *AuthS) Logout(ctx context.Context, tokenID string) error {
	if err := a.repo.DeleteToken(ctx, tokenID); err != nil {
		a.log.Error("failed to delete refresh token during logout",
			zap.String("token_id", tokenID),
			zap.Error(err),
		)
		return err
	}

	a.log.Info("user logged out successfully",
		zap.String("token_id", tokenID),
	)

	return nil
}

func (a *AuthS) checkUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	userID, hashedPass, err := a.repo.UserCredentials(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			a.log.Warn("login attempt with non-existent email",
				zap.String("email", email),
			)
		} else {
			a.log.Error("failed to get user credentials from repo",
				zap.String("email", email),
				zap.Error(err),
			)
		}
		return uuid.Nil, err
	}

	if err := a.hasher.ComparePassword(hashedPass, password); err != nil {
		a.log.Warn("incorrect password provided",
			zap.String("email", email),
			zap.String("user_id", userID.String()),
		)
		return uuid.Nil, domain.ErrIncorrectPassword
	}

	return userID, nil
}

func (a *AuthS) generateAndSaveTokens(ctx context.Context, userID uuid.UUID) (dto.TokenOutput, error) {
	accessToken, refreshToken, err := a.generateTokens(userID)
	if err != nil {
		return dto.TokenOutput{}, err
	}

	if err := a.repo.CreateToken(ctx, refreshToken); err != nil {
		return dto.TokenOutput{}, err
	}

	return dto.TokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.TokenID,
	}, nil
}

func (a *AuthS) generateTokens(userID uuid.UUID) (string, domain.Token, error) {
	accessToken, err := a.generateAccessToken(userID)
	if err != nil {
		return "", domain.Token{}, err
	}

	refreshToken := a.generateRefreshToken(userID)
	return accessToken, refreshToken, nil
}

func (a *AuthS) generateAccessToken(userID uuid.UUID) (string, error) {
	tkn := jwt.New()
	if err := tkn.Set(jwt.SubjectKey, userID.String()); err != nil {
		return "", fmt.Errorf("failed to set subject in token: %w", err)
	}

	if err := tkn.Set(jwt.ExpirationKey, time.Now().Add(a.token.AccessTokenTTL)); err != nil {
		return "", fmt.Errorf("failed to set expiration in token: %w", err)
	}

	if err := tkn.Set(jwt.IssuedAtKey, time.Now()); err != nil {
		return "", fmt.Errorf("failed to set issued at in token: %w", err)
	}

	accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte(a.token.JwtSecret)))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %s", err)
	}

	return string(accessToken), nil
}

func (a *AuthS) generateRefreshToken(userID uuid.UUID) domain.Token {
	return domain.Token{
		UserID:    userID,
		TokenID:   uuid.New().String(),
		ExpiresAt: time.Now().Add(a.token.RefreshTokenTTL),
	}
}

func (a *AuthS) ParseToken(ctx context.Context, accessToken string) (uuid.UUID, error) {
	verified, err := jwt.Parse([]byte(accessToken), jwt.WithKey(jwa.HS256, []byte(a.token.JwtSecret)))
	if err != nil {
		a.log.Debug("failed to parse or verify access token",
			zap.Error(err),
		)
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	subject, ok := verified.Get(jwt.SubjectKey)
	if !ok {
		a.log.Debug("token missing 'sub' claim")
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	subjectStr, ok := subject.(string)
	if !ok {
		a.log.Debug("token 'sub' claim is not a string")
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(subjectStr)
	if err != nil {
		a.log.Debug("invalid user ID in token",
			zap.String("subject", subjectStr),
			zap.Error(err),
		)
		return uuid.Nil, domain.ErrInvalidUUID
	}

	return userID, nil
}

func (a *AuthS) RefreshToken(ctx context.Context, tokenID string) (dto.TokenOutput, error) {
	tokenDB, err := a.repo.Token(ctx, tokenID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			a.log.Warn("refresh token not found (possible reuse or logout)",
				zap.String("token_id", tokenID),
			)
		} else {
			a.log.Error("failed to get refresh token from repo",
				zap.String("token_id", tokenID),
				zap.Error(err),
			)
		}
		return dto.TokenOutput{}, err
	}

	if err := a.repo.DeleteToken(ctx, tokenID); err != nil {
		a.log.Error("failed to delete old refresh token",
			zap.String("token_id", tokenID),
			zap.Error(err),
		)
		return dto.TokenOutput{}, err
	}

	if tokenDB.ExpiresAt.Before(time.Now()) {
		a.log.Warn("attempt to refresh expired token",
			zap.String("token_id", tokenID),
			zap.Time("expires_at", tokenDB.ExpiresAt),
		)
		return dto.TokenOutput{}, fmt.Errorf("token expired")
	}

	token, err := a.generateAndSaveTokens(ctx, tokenDB.UserID)
	if err != nil {
		a.log.Error("failed to generate new tokens during refresh",
			zap.String("user_id", tokenDB.UserID.String()),
			zap.Error(err),
		)
		return dto.TokenOutput{}, err
	}

	a.log.Info("token refreshed successfully",
		zap.String("user_id", tokenDB.UserID.String()),
		zap.String("old_token_id", tokenID),
		zap.String("new_refresh_token_id", token.RefreshToken),
	)

	return token, nil
}
