package repository

import (
	"context"
	"noteApp/internal/models/domain"
	"noteApp/pkg/logger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserR_CreateUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		user domain.User
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
				user: domain.User{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate username",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					ID:       uuid.New(),
					Username: "duplicate_test_username",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "duplicate email",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "duplicate_test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty user",
			args: args{
				ctx:  context.Background(),
				user: domain.User{},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty username",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					ID:       uuid.New(),
					Username: "",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty email",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "",
					Password: "test_password",
					ImageURL: "test_image",
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty password",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    "test_email",
					Password: "",
					ImageURL: "test_image",
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

			repo := NewUserRepository(tx, logger.LoggerForTest())

			t.Cleanup(func() { _ = tx.Rollback() })

			switch tt.name {
			case "duplicate username":
				err = repo.CreateUser(tt.args.ctx, domain.User{
					ID:       uuid.New(),
					Username: "duplicate_test_username",
					Email:    "test_email1",
					Password: "test_password1",
					ImageURL: "test_image1",
				})
				require.NoError(t, err)
			case "duplicate email":
				err = repo.CreateUser(tt.args.ctx, domain.User{
					ID:       uuid.New(),
					Username: "test_username1",
					Email:    "duplicate_test_email",
					Password: "test_password1",
					ImageURL: "test_image1",
				})
				require.NoError(t, err)
			default:
			}

			err = repo.CreateUser(tt.args.ctx, tt.args.user)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)

			user, err := repo.UserByID(tt.args.ctx, tt.args.user.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.args.user.ID, user.ID)
			assert.Equal(t, tt.args.user.Username, user.Username)
			assert.Equal(t, tt.args.user.Email, user.Email)
			assert.Equal(t, tt.args.user.ImageURL, user.ImageURL)
			require.WithinDuration(t, time.Now().UTC(), user.CreatedAt, time.Second)
			require.WithinDuration(t, time.Now().UTC(), user.UpdatedAt, time.Second)
		})
	}
}

func TestUserR_UserByID(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
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
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
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

			repo := NewUserRepository(tx, logger.LoggerForTest())

			testUser := domain.User{}

			if tt.name == "success" {
				testUser = domain.User{
					ID:       tt.args.userID,
					Username: "test_username",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				}
				err = repo.CreateUser(tt.args.ctx, testUser)
				require.NoError(t, err)
			}

			got, err := repo.UserByID(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, got.ID, testUser.ID)
			assert.Equal(t, got.Username, testUser.Username)
			assert.Equal(t, got.Email, testUser.Email)
			assert.Equal(t, got.ImageURL, testUser.ImageURL)
			require.WithinDuration(t, time.Now().UTC(), got.CreatedAt, time.Second)
			require.WithinDuration(t, time.Now().UTC(), got.UpdatedAt, time.Second)
		})
	}
}

func TestUserR_UserCredentials(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		email string
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
				ctx:   context.Background(),
				email: "test_email",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			args: args{
				ctx:   context.Background(),
				email: "invalid_email",
			},
			wantErr:    true,
			wantErrMsg: domain.ErrNotFound.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)

			t.Cleanup(func() { _ = tx.Rollback() })

			repo := NewUserRepository(tx, logger.LoggerForTest())

			testUser := domain.User{}

			if tt.name == "success" {
				testUser = domain.User{
					ID:       uuid.New(),
					Username: "test_username",
					Email:    tt.args.email,
					Password: "test_password",
					ImageURL: "test_image",
				}
				err = repo.CreateUser(tt.args.ctx, testUser)
				require.NoError(t, err)
			}

			got, got1, err := repo.UserCredentials(tt.args.ctx, tt.args.email)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testUser.ID, got)
			assert.Equal(t, testUser.Password, got1)
		})
	}
}

func TestUserR_UpdateUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		user domain.UserUpdate
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
				user: domain.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"updated_username"}[0],
					Email:    &[]string{"updated_email"}[0],
					Password: &[]string{"updated_password"}[0],
					ImageURL: &[]string{"updated_image"}[0],
				},
			},
			wantErr: false,
		},
		{
			name: "empty fields",
			args: args{
				ctx: context.Background(),
				user: domain.UserUpdate{
					ID: uuid.New(),
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrNoFieldsToUpdate.Error(),
		},
		{
			name: "empty username",
			args: args{
				ctx: context.Background(),
				user: domain.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{""}[0],
					Email:    &[]string{"updated_email"}[0],
					Password: &[]string{"updated_password"}[0],
					ImageURL: &[]string{"updated_image"}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
		{
			name: "empty email",
			args: args{
				ctx: context.Background(),
				user: domain.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"updated_username"}[0],
					Email:    &[]string{""}[0],
					Password: &[]string{"updated_password"}[0],
					ImageURL: &[]string{"updated_image"}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
		{
			name: "empty password",
			args: args{
				ctx: context.Background(),
				user: domain.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"updated_username"}[0],
					Email:    &[]string{"updated_email"}[0],
					Password: &[]string{""}[0],
					ImageURL: &[]string{"updated_image"}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
		{
			name: "duplicate username",
			args: args{
				ctx: context.Background(),
				user: domain.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"duplicate_username"}[0],
					Email:    &[]string{"updated_email"}[0],
					Password: &[]string{"updated_password"}[0],
					ImageURL: &[]string{"updated_image"}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
		{
			name: "duplicate email",
			args: args{
				ctx: context.Background(),
				user: domain.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"updated_username"}[0],
					Email:    &[]string{"duplicate_email"}[0],
					Password: &[]string{"updated_password"}[0],
					ImageURL: &[]string{"updated_image"}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)

			t.Cleanup(func() { _ = tx.Rollback() })

			repo := NewUserRepository(tx, logger.LoggerForTest())

			testUser := domain.User{
				ID:       tt.args.user.ID,
				Username: "test_username",
				Email:    "test_email",
				Password: "test_password",
				ImageURL: "test_image",
			}

			switch tt.name {
			case "duplicate username":
				err = repo.CreateUser(tt.args.ctx, domain.User{
					ID:       uuid.New(),
					Username: "duplicate_username",
					Email:    "test_email1",
					Password: "test_password1",
					ImageURL: "test_image1",
				})
				require.NoError(t, err)
			case "duplicate email":
				err = repo.CreateUser(tt.args.ctx, domain.User{
					ID:       uuid.New(),
					Username: "test_username1",
					Email:    "duplicate_email",
					Password: "test_password1",
					ImageURL: "test_image1",
				})
				require.NoError(t, err)
			default:
			}

			err = repo.CreateUser(tt.args.ctx, testUser)
			require.NoError(t, err)

			err = repo.UpdateUser(tt.args.ctx, tt.args.user)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			got, err := repo.UserByID(tt.args.ctx, tt.args.user.ID)
			require.NoError(t, err)
			if tt.args.user.Username != nil {
				assert.Equal(t, *tt.args.user.Username, got.Username)
			} else {
				assert.Equal(t, testUser.Username, got.Username)
			}

			if tt.args.user.Email != nil {
				assert.Equal(t, *tt.args.user.Email, got.Email)
			} else {
				assert.Equal(t, testUser.Email, got.Email)
			}

			if tt.args.user.ImageURL != nil {
				assert.Equal(t, *tt.args.user.ImageURL, got.ImageURL)
			} else {
				assert.Equal(t, testUser.ImageURL, got.ImageURL)
			}

			if tt.args.user.Password != nil {
				_, got1, err := repo.UserCredentials(tt.args.ctx, *tt.args.user.Email)
				require.NoError(t, err)
				assert.Equal(t, *tt.args.user.Password, got1)
			}

			require.WithinDuration(t, time.Now().UTC(), got.UpdatedAt, time.Second)
			require.Greater(t, got.UpdatedAt.Unix(), testUser.CreatedAt.Unix())
		})
	}
}

func TestUserR_DeleteUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
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
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			wantErr:    true,
			wantErrMsg: domain.ErrNotFound.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)

			t.Cleanup(func() { _ = tx.Rollback() })

			repo := NewUserRepository(tx, logger.LoggerForTest())

			if tt.name == "success" {
				testUser := domain.User{
					ID:       tt.args.userID,
					Username: "test_username",
					Email:    "test_email",
					Password: "test_password",
					ImageURL: "test_image",
				}
				err = repo.CreateUser(tt.args.ctx, testUser)
				require.NoError(t, err)
			}

			err = repo.DeleteUser(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)

			_, err = repo.UserByID(tt.args.ctx, tt.args.userID)
			require.Error(t, err)
		})
	}
}
