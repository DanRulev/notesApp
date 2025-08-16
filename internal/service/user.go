package service

import (
	"context"
	"errors"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserRI interface {
	UserByID(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.UserUpdate) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type UserCred interface {
	UserCredentials(ctx context.Context, email string) (uuid.UUID, string, error)
}

type UserS struct {
	repo   UserRI
	cred   UserCred
	hasher HasherI
	log    *logger.Logger
}

func NewUserService(
	repo UserRI,
	cred UserCred,
	hasher HasherI,
	log *logger.Logger,
) *UserS {
	return &UserS{
		repo:   repo,
		cred:   cred,
		hasher: hasher,
		log:    log,
	}
}

func (u *UserS) UpdateUserPassword(ctx context.Context, updPass dto.UserUpdPassword) error {
	userDB, err := u.repo.UserByID(ctx, updPass.UserID)
	if err != nil {
		u.log.Error("failed to get user by ID during password update",
			zap.Error(err),
			zap.String("user_id", updPass.UserID.String()),
		)
		return err
	}

	userID, oldPassword, _ := u.cred.UserCredentials(ctx, userDB.Email)

	if err := u.hasher.ComparePassword(oldPassword, updPass.OldPassword); err != nil {
		u.log.Warn("incorrect old password provided",
			zap.String("user_id", userID.String()),
			zap.String("email", userDB.Email),
		)
		return domain.ErrIncorrectPassword
	}

	hashedPassword, err := u.hasher.GenerateHash(updPass.NewPassword)
	if err != nil {
		u.log.Error("failed to hash new password",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return err
	}

	inputUser := domain.UserUpdate{
		ID:       userID,
		Password: &hashedPassword,
	}

	if err := u.repo.UpdateUser(ctx, inputUser); err != nil {
		u.log.Error("failed to update user in repo",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)

		return err
	}

	u.log.Info("user password updated successfully",
		zap.String("user_id", userID.String()),
	)

	return nil
}

func (u *UserS) UserByID(ctx context.Context, id uuid.UUID) (dto.UserOutput, error) {
	userDB, err := u.repo.UserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			u.log.Warn("user not found by ID",
				zap.String("user_id", id.String()),
			)
		} else {
			u.log.Error("failed to get user from repository",
				zap.Error(err),
				zap.String("user_id", id.String()),
			)
		}
		return dto.UserOutput{}, err
	}

	u.log.Debug("user retrieved from DB",
		zap.String("user_id", userDB.ID.String()),
		zap.String("email", userDB.Email),
	)

	return dto.UserOutput{
		ID:        userDB.ID,
		Username:  userDB.Username,
		Email:     userDB.Email,
		ImageURL:  userDB.ImageURL,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
	}, nil
}

func (u *UserS) UpdateUser(ctx context.Context, user dto.UserUpdate) error {
	input := domain.UserUpdate{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		ImageURL: user.ImageURL,
	}

	if err := input.Validate(); err != nil {
		u.log.Debug("user update validation failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return err
	}

	if err := u.repo.UpdateUser(ctx, input); err != nil {
		if err == domain.ErrNotFound {
			u.log.Warn("user not found",
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
		} else {
			u.log.Error("failed to update user in repo",
				zap.Error(err),
				zap.String("user_id", user.ID.String()),
			)
		}
		return err
	}

	u.log.Info("user updated successfully",
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

func (u *UserS) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			u.log.Warn("user not found during deletion",
				zap.String("user_id", id.String()),
			)
		} else {
			u.log.Error("failed to delete user",
				zap.Error(err),
				zap.String("user_id", id.String()),
			)
		}
		return err
	}

	u.log.Info("user deleted successfully",
		zap.String("user_id", id.String()),
	)

	return nil
}
