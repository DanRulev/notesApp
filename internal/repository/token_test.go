package repository

import (
	"context"
	"noteApp/internal/models/domain"
	"noteApp/pkg/logger"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenR_CreateToken(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		token domain.Token
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				token: domain.Token{
					UserID:    uuid.New(),
					TokenID:   uuid.New().String(),
					ExpiresAt: time.Now().Add(5 * time.Second).UTC(),
				},
			},
			wantErr: false,
		},
		{
			name: "without user",
			args: args{
				ctx: context.Background(),
				token: domain.Token{
					UserID:    uuid.Nil,
					TokenID:   uuid.New().String(),
					ExpiresAt: time.Now().Add(5 * time.Second).UTC(),
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty token ID",
			args: args{
				ctx: context.Background(),
				token: domain.Token{
					UserID:    uuid.New(),
					TokenID:   "",
					ExpiresAt: time.Now().Add(5 * time.Second),
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = tx.Rollback() })

			if tt.name != "without user" {
				repoUser := NewUserRepository(tx, logger.LoggerForTest())
				err = repoUser.CreateUser(context.Background(), domain.User{
					ID:       tt.args.token.UserID,
					Username: "test",
					Email:    "test",
					Password: "test",
					ImageURL: "test",
				})

				require.NoError(t, err)
			}

			repo := NewRepository(tx, logger.LoggerForTest())

			err = repo.CreateToken(tt.args.ctx, tt.args.token)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			token, err := repo.Token(tt.args.ctx, tt.args.token.TokenID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, token.UserID, tt.args.token.UserID)
			require.WithinDuration(t, tt.args.token.ExpiresAt, token.ExpiresAt, time.Second)
		})
	}
}

func TestTokenR_Token(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		tokenID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			wantErr: false,
		},
		{
			name: "invalid token ID",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			wantErr:    true,
			wantErrMsg: domain.ErrReceiving.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = tx.Rollback() })

			repo := NewRepository(tx, logger.LoggerForTest())

			var testToken domain.Token

			switch tt.name {
			case "success":
				userID := uuid.New()
				repoUser := NewUserRepository(tx, logger.LoggerForTest())
				err = repoUser.CreateUser(context.Background(), domain.User{
					ID:       userID,
					Username: "test",
					Email:    "test",
					Password: "test",
					ImageURL: "test",
				})

				require.NoError(t, err)

				testToken = domain.Token{
					UserID:    userID,
					TokenID:   tt.args.tokenID,
					ExpiresAt: time.Now().Add(5 * time.Second).Local().UTC(),
				}
				err = repo.CreateToken(tt.args.ctx, testToken)
				require.NoError(t, err)
			default:
			}

			got, err := repo.Token(tt.args.ctx, tt.args.tokenID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testToken.UserID, got.UserID)
			assert.Equal(t, testToken.TokenID, got.TokenID)
			require.WithinDuration(t, testToken.ExpiresAt, got.ExpiresAt, time.Second)
		})
	}
}

func TestTokenR_DeleteToken(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		tokenID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			wantErr: false,
		},
		{
			name: "invalid token ID",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New().String(),
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToDelete.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = tx.Rollback() })

			repo := NewRepository(tx, logger.LoggerForTest())

			switch tt.name {
			case "success":
				userID := uuid.New()
				repoUser := NewUserRepository(tx, logger.LoggerForTest())
				err = repoUser.CreateUser(context.Background(), domain.User{
					ID:       userID,
					Username: "test",
					Email:    "test",
					Password: "test",
					ImageURL: "test",
				})

				require.NoError(t, err)

				err = repo.CreateToken(tt.args.ctx, domain.Token{
					UserID:    userID,
					TokenID:   tt.args.tokenID,
					ExpiresAt: time.Now().Add(5 * time.Second),
				})
				require.NoError(t, err)
			default:
			}

			err = repo.DeleteToken(tt.args.ctx, tt.args.tokenID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)

			_, err = repo.Token(tt.args.ctx, tt.args.tokenID)
			require.Error(t, err)
		})
	}
}
