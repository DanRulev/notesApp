package service

import (
	"context"
	"errors"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	mock_service "noteApp/internal/service/mock"
	"noteApp/pkg/logger"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockUserService(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_service.MockRepositoryI, *mock_service.MockHasherI)) *UserS {
	t.Helper()

	repo := mock_service.NewMockRepositoryI(ctrl)
	hasher := mock_service.NewMockHasherI(ctrl)
	if setupMock != nil {
		setupMock(repo, hasher)
	}

	return NewUserService(repo, repo, hasher, logger.LoggerForTest())
}

func TestUserS_UpdateUserPassword(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		updPass dto.UserUpdPassword
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
				updPass: dto.UserUpdPassword{
					UserID:      uuid.New(),
					OldPassword: "old_password",
					NewPassword: "new_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, nil)
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.New(), "old_password", nil)
				mhi.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
				mhi.EXPECT().GenerateHash(gomock.Any()).Return("hashed_password", nil)
				mri.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				updPass: dto.UserUpdPassword{
					UserID:      uuid.New(),
					OldPassword: "old_password",
					NewPassword: "new_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("repository error"))
			},
			wantErr: true,
		},
		{
			name: "wrong old password",
			args: args{
				ctx: context.Background(),
				updPass: dto.UserUpdPassword{
					UserID:      uuid.New(),
					OldPassword: "old_password",
					NewPassword: "new_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, nil)
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.UUID{}, "old_password", nil)
				mhi.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(errors.New(""))
			},
			wantErr: true,
		},
		{
			name: "empty new password",
			args: args{
				ctx: context.Background(),
				updPass: dto.UserUpdPassword{
					UserID:      uuid.New(),
					OldPassword: "old_password",
					NewPassword: "",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, nil)
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.UUID{}, "old_password", nil)
				mhi.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
				mhi.EXPECT().GenerateHash(gomock.Any()).Return("hashed_password", errors.New("empty password"))
			},
			wantErr: true,
		},
		{
			name: "failed to update",
			args: args{
				ctx: context.Background(),
				updPass: dto.UserUpdPassword{
					UserID:      uuid.New(),
					OldPassword: "old_password",
					NewPassword: "new_password",
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, nil)
				mri.EXPECT().UserCredentials(gomock.Any(), gomock.Any()).Return(uuid.New(), "old_password", nil)
				mhi.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
				mhi.EXPECT().GenerateHash(gomock.Any()).Return("hashed_password", nil)
				mri.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToUpdate)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockUserService(t, ctrl, tt.f)

			err := repo.UpdateUserPassword(tt.args.ctx, tt.args.updPass)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserS_UserByID(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  uuid.UUID
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
				id:  uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{
					ID:        uuid.New(),
					Username:  "username",
					Email:     "email",
					ImageURL:  "image",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			args: args{
				ctx: context.Background(),
				id:  uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				id:  uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockUserService(t, ctrl, tt.f)

			got, err := repo.UserByID(tt.args.ctx, tt.args.id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, got.ID, uuid.Nil)
			assert.NotEqual(t, got.Username, "")
			assert.NotEqual(t, got.Email, "")
			assert.NotEqual(t, got.ImageURL, "")
			assert.True(t, !got.CreatedAt.IsZero())
			assert.True(t, !got.UpdatedAt.IsZero())
		})
	}
}

func TestUserS_UpdateUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		user dto.UserUpdate
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
				user: dto.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"username"}[0],
					Email:    &[]string{"email"}[0],
					ImageURL: &[]string{"image"}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx: context.Background(),
				user: dto.UserUpdate{
					ID:       uuid.Nil,
					Username: &[]string{"username"}[0],
					Email:    &[]string{"email"}[0],
					ImageURL: &[]string{"image"}[0],
				},
			},
			wantErr: true,
		},
		{
			name: "invalid username",
			args: args{
				ctx: context.Background(),
				user: dto.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{""}[0],
					Email:    &[]string{"email"}[0],
					ImageURL: &[]string{"image"}[0],
				},
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			args: args{
				ctx: context.Background(),
				user: dto.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"username"}[0],
					Email:    &[]string{""}[0],
					ImageURL: &[]string{"image"}[0],
				},
			},
			wantErr: true,
		},
		{
			name: "update error",
			args: args{
				ctx: context.Background(),
				user: dto.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"username"}[0],
					Email:    &[]string{"email"}[0],
					ImageURL: &[]string{"image"}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "update error",
			args: args{
				ctx: context.Background(),
				user: dto.UserUpdate{
					ID:       uuid.New(),
					Username: &[]string{"username"}[0],
					Email:    &[]string{"email"}[0],
					ImageURL: &[]string{"image"}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToUpdate)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockUserService(t, ctrl, tt.f)

			err := repo.UpdateUser(tt.args.ctx, tt.args.user)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserS_DeleteUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  uuid.UUID
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
				id:  uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				ctx: context.Background(),
				id:  uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "delete error",
			args: args{
				ctx: context.Background(),
				id:  uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI, mhi *mock_service.MockHasherI) {
				mri.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(domain.ErrFailedToDelete)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockUserService(t, ctrl, tt.f)

			err := repo.DeleteUser(tt.args.ctx, tt.args.id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
