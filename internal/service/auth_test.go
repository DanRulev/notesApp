package service

import (
	"context"
	"errors"
	"fmt"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	mock_service "noteApp/internal/service/mock"
	"noteApp/pkg/logger"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockAuthService(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)) *AuthS {
	t.Helper()

	repo := mock_service.NewMockRepositoryI(ctrl)
	hasher := mock_service.NewMockHasherI(ctrl)
	if setupMock != nil {
		setupMock(repo, hasher)
	}

	cfg, err := initConfig()
	require.NoError(t, err)

	return NewAuthService(repo, hasher, cfg, logger.LoggerForTest())
}

func TestAuthS_SignUp(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		user dto.UserCreate
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				user: dto.UserCreate{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				hasher.EXPECT().GenerateHash(gomock.Any()).Return("hashed", nil)
				mri.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty username",
			args: args{
				ctx: context.Background(),
				user: dto.UserCreate{
					ID:       uuid.New(),
					Username: "",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				hasher.EXPECT().GenerateHash(gomock.Any()).Return("hashed", nil)
			},
			wantErr: true,
		},
		{
			name: "empty email",
			args: args{
				ctx: context.Background(),
				user: dto.UserCreate{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				hasher.EXPECT().GenerateHash(gomock.Any()).Return("hashed", nil)
			},
			wantErr: true,
		},
		{
			name: "empty password",
			args: args{
				ctx: context.Background(),
				user: dto.UserCreate{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "test_email",
					Password: "",
					ImageURL: "test_image",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				hasher.EXPECT().GenerateHash(gomock.Any()).Return("", errors.New("empty password"))

			},
			wantErr: true,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				user: dto.UserCreate{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				hasher.EXPECT().GenerateHash(gomock.Any()).Return("hashed", nil)
				mri.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockAuthService(t, ctrl, tt.f)

			got, err := repo.SignUp(tt.args.ctx, tt.args.user)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, got != uuid.Nil)
		})
	}
}

func TestAuthS_SignIn(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		data dto.UserSignIn
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				data: dto.UserSignIn{
					Email:    "test_email",
					Password: "test_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.UUID{}, "hashed_pass", nil)
				hasher.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
				mri.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty data",
			args: args{
				ctx: context.Background(),
				data: dto.UserSignIn{
					Email:    "",
					Password: "",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.Nil, "", domain.ErrReceiving)
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			args: args{
				ctx: context.Background(),
				data: dto.UserSignIn{
					Email:    "wrong_email",
					Password: "test_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.Nil, "", domain.ErrReceiving)
			},
			wantErr: true,
		},
		{
			name: "user not found",
			args: args{
				ctx: context.Background(),
				data: dto.UserSignIn{
					Email:    "test_email",
					Password: "test_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.UUID{}, "", domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "invalid password",
			args: args{
				ctx: context.Background(),
				data: dto.UserSignIn{
					Email:    "test_email",
					Password: "wrong_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.UUID{}, "hashed_pass", nil)
				hasher.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(errors.New("invalid password"))
			},
			wantErr: true,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				data: dto.UserSignIn{
					Email:    "test_email",
					Password: "test_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, hasher *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.UUID{}, "hashed_pass", nil)
				hasher.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
				mri.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got, err := a.SignIn(tt.args.ctx, tt.args.data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, got.AccessToken != "")
			assert.True(t, got.RefreshToken != "")
		})
	}
}

func TestAuthS_Logout(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		tokenID string
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid token ID",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToDelete)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			err := a.Logout(tt.args.ctx, tt.args.tokenID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAuthS_checkUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	type args struct {
		ctx      context.Context
		email    string
		password string
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:      context.Background(),
				email:    "test_email",
				password: "test_password",
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(userID, "hashed", nil)
				mhi.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			args: args{
				ctx:      context.Background(),
				email:    "invalid",
				password: "test_password",
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.Nil, "", domain.ErrReceiving)
			},
			wantErr: true,
		},
		{
			name: "invalid password",
			args: args{
				ctx:      context.Background(),
				email:    "test_email",
				password: "test_password",
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(userID, "hashed", nil)
				mhi.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(errors.New(""))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got, err := a.checkUser(tt.args.ctx, tt.args.email, tt.args.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, got, userID)
		})
	}
}

func TestAuthS_generateAndSaveTokens(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToCreate)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got, err := a.generateAndSaveTokens(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, got)
			assert.NotEqual(t, got.AccessToken, uuid.Nil.String())
			assert.NotEqual(t, got.RefreshToken, uuid.Nil.String())
		})
	}
}

func TestAuthS_generateTokens(t *testing.T) {
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				userID: uuid.New(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got1, got2, err := a.generateTokens(tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, got1, uuid.Nil.String())
			assert.NotEqual(t, got2, uuid.Nil.String())
		})
	}
}

func TestAuthS_generateAccessToken(t *testing.T) {
	t.Parallel()

	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    string
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				userID: uuid.New(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got, err := a.generateAccessToken(tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEqual(t, got, uuid.Nil.String())
		})
	}
}

func TestAuthS_generateRefreshToken(t *testing.T) {
	t.Parallel()

	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name string
		args args
		f    func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
	}{
		{
			name: "success",
			args: args{
				userID: uuid.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got := a.generateRefreshToken(tt.args.userID)

			require.Equal(t, got.UserID, tt.args.userID)
			assert.NotEqual(t, got.TokenID, uuid.Nil.String())
			assert.WithinDuration(t, got.ExpiresAt, time.Now().Add(a.token.RefreshTokenTTL), time.Second)
		})
	}
}

func TestAuthS_ParseToken(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		f        func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		generate func(uuid.UUID, time.Duration, string) (string, error)
		wantErr  bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			generate: func(u uuid.UUID, t time.Duration, s string) (string, error) {
				tkn := jwt.New()
				if err := tkn.Set(jwt.SubjectKey, u.String()); err != nil {
					return "", fmt.Errorf("failed to set subject in token: %w", err)
				}

				if err := tkn.Set(jwt.ExpirationKey, time.Now().Add(t)); err != nil {
					return "", fmt.Errorf("failed to set expiration in token: %w", err)
				}

				if err := tkn.Set(jwt.IssuedAtKey, time.Now()); err != nil {
					return "", fmt.Errorf("failed to set issued at in token: %w", err)
				}

				accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte(s)))
				if err != nil {
					return "", fmt.Errorf("failed to sign token: %s", err)
				}

				return string(accessToken), nil
			},
			wantErr: false,
		},
		{
			name: "wrong secret",
			args: args{
				ctx: context.Background(),
			},
			generate: func(u uuid.UUID, t time.Duration, s string) (string, error) {
				tkn := jwt.New()
				if err := tkn.Set(jwt.SubjectKey, u.String()); err != nil {
					return "", fmt.Errorf("failed to set subject in token: %w", err)
				}

				if err := tkn.Set(jwt.ExpirationKey, time.Now().Add(t)); err != nil {
					return "", fmt.Errorf("failed to set expiration in token: %w", err)
				}

				if err := tkn.Set(jwt.IssuedAtKey, time.Now()); err != nil {
					return "", fmt.Errorf("failed to set issued at in token: %w", err)
				}

				accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte(s+"wrong")))
				if err != nil {
					return "", fmt.Errorf("failed to sign token: %s", err)
				}

				return string(accessToken), nil
			},
			wantErr: true,
		},
		{
			name: "without subject",
			args: args{
				ctx: context.Background(),
			},
			generate: func(u uuid.UUID, t time.Duration, s string) (string, error) {
				tkn := jwt.New()
				if err := tkn.Set(jwt.ExpirationKey, time.Now().Add(t)); err != nil {
					return "", fmt.Errorf("failed to set expiration in token: %w", err)
				}

				if err := tkn.Set(jwt.IssuedAtKey, time.Now()); err != nil {
					return "", fmt.Errorf("failed to set issued at in token: %w", err)
				}

				accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte(s)))
				if err != nil {
					return "", fmt.Errorf("failed to sign token: %s", err)
				}

				return string(accessToken), nil
			},
			wantErr: true,
		},
		{
			name: "wrong subject type",
			args: args{
				ctx: context.Background(),
			},
			generate: func(u uuid.UUID, t time.Duration, s string) (string, error) {
				tkn := jwt.New()
				if err := tkn.Set(jwt.SubjectKey, 6); err != nil {
					return "", fmt.Errorf("failed to set subject in token: %w", err)
				}

				if err := tkn.Set(jwt.ExpirationKey, time.Now().Add(t)); err != nil {
					return "", fmt.Errorf("failed to set expiration in token: %w", err)
				}

				if err := tkn.Set(jwt.IssuedAtKey, time.Now()); err != nil {
					return "", fmt.Errorf("failed to set issued at in token: %w", err)
				}

				accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte(s)))
				if err != nil {
					return "", fmt.Errorf("failed to sign token: %s", err)
				}

				return string(accessToken), nil
			},
			wantErr: true,
		},
		{
			name: "subject not uuid",
			args: args{
				ctx: context.Background(),
			},
			generate: func(u uuid.UUID, t time.Duration, s string) (string, error) {
				tkn := jwt.New()
				if err := tkn.Set(jwt.SubjectKey, "wrong"); err != nil {
					return "", fmt.Errorf("failed to set subject in token: %w", err)
				}

				if err := tkn.Set(jwt.ExpirationKey, time.Now().Add(t)); err != nil {
					return "", fmt.Errorf("failed to set expiration in token: %w", err)
				}

				if err := tkn.Set(jwt.IssuedAtKey, time.Now()); err != nil {
					return "", fmt.Errorf("failed to set issued at in token: %w", err)
				}

				accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte(s)))
				if err != nil {
					return "", fmt.Errorf("failed to sign token: %s", err)
				}

				return string(accessToken), nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			userID := uuid.New()
			accessToken, err := tt.generate(userID, a.token.AccessTokenTTL, a.token.JwtSecret)
			require.NoError(t, err)

			got, err := a.ParseToken(tt.args.ctx, accessToken)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, got, userID)
		})
	}
}

func TestAuthS_RefreshToken(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		tokenID string
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().Token(gomock.Any(), gomock.Any()).Return(domain.Token{
					UserID:    uuid.New(),
					TokenID:   uuid.New().String(),
					ExpiresAt: time.Now().Add(5 * time.Hour),
				}, nil)
				mri.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).Return(nil)
				mri.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "token not found",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().Token(gomock.Any(), gomock.Any()).Return(domain.Token{}, domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().Token(gomock.Any(), gomock.Any()).Return(domain.Token{}, errors.New("repository error"))
			},
			wantErr: true,
		},
		{
			name: "failed delete token",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().Token(gomock.Any(), gomock.Any()).Return(domain.Token{
					UserID:    uuid.New(),
					TokenID:   uuid.New().String(),
					ExpiresAt: time.Now().Add(5 * time.Hour),
				}, nil)
				mri.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToDelete)
			},
			wantErr: true,
		},
		{
			name: "token expired",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().Token(gomock.Any(), gomock.Any()).Return(domain.Token{
					UserID:    uuid.New(),
					TokenID:   uuid.New().String(),
					ExpiresAt: time.Now().Add(-5 * time.Hour),
				}, nil)
				mri.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "failed save token",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().Token(gomock.Any(), gomock.Any()).Return(domain.Token{
					UserID:    uuid.New(),
					TokenID:   uuid.New().String(),
					ExpiresAt: time.Now().Add(5 * time.Hour),
				}, nil)
				mri.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).Return(nil)
				mri.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToCreate)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := mockAuthService(t, ctrl, tt.f)

			got, err := a.RefreshToken(tt.args.ctx, tt.args.tokenID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, got.AccessToken, uuid.Nil)
			assert.NotEqual(t, got.RefreshToken, uuid.Nil.String())
		})
	}
}
