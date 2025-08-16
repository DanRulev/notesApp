package repository

import (
	"context"
	"database/sql"
	"fmt"
	"noteApp/internal/models/domain"
	"noteApp/pkg/logger"
	"noteApp/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserR struct {
	db  query
	log *logger.Logger
}

func NewUserRepository(db query, log *logger.Logger) *UserR {
	return &UserR{
		db:  db,
		log: log,
	}
}

func (u *UserR) CreateUser(ctx context.Context, user domain.User) error {
	query := `INSERT INTO users (id, username, email, password, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := u.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.Password, user.ImageURL, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		u.log.Error("failed to execute INSERT query in CreateUser",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email),
		)
		return domain.MakeError(domain.ErrFailedToCreate, err, "user")
	}

	return nil
}

func (u *UserR) UserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	query := `SELECT id, username, email, image_url, created_at, updated_at FROM users WHERE id = $1`

	var user domain.User
	row := u.db.QueryRowContext(ctx, query, userID)
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.ImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, domain.MakeError(domain.ErrReceiving, domain.ErrNotFound, "user")
		}
		u.log.Error("database error in UserByID query",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return domain.User{}, domain.MakeError(domain.ErrReceiving, err, "user")
	}

	return user, nil
}

func (u *UserR) UserCredentials(ctx context.Context, email string) (uuid.UUID, string, error) {
	query := `SELECT id, password FROM users WHERE email = $1`

	var (
		password string
		userID   uuid.UUID
	)
	err := u.db.QueryRowContext(ctx, query, email).Scan(&userID, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, "", domain.MakeError(domain.ErrReceiving, domain.ErrNotFound, "user")
		}
		u.log.Error("database error in UserCredentials query",
			zap.Error(err),
			zap.String("email", email),
		)
		return uuid.Nil, "", domain.MakeError(domain.ErrReceiving, err, "user")
	}

	return userID, password, nil
}

func (u *UserR) UpdateUser(ctx context.Context, user domain.UserUpdate) error {
	var (
		fields []string
		args   []interface{}
		argIdx int
	)

	utils.AddFieldsToQuery("username", user.Username, &fields, &args, &argIdx)
	utils.AddFieldsToQuery("email", user.Email, &fields, &args, &argIdx)
	utils.AddFieldsToQuery("password", user.Password, &fields, &args, &argIdx)
	utils.AddFieldsToQuery("image_url", user.ImageURL, &fields, &args, &argIdx)

	if len(fields) == 0 {
		return domain.MakeError(domain.ErrFailedToUpdate, domain.ErrNoFieldsToUpdate, "user")
	}

	fields = append(fields, "updated_at=NOW()")

	query := fmt.Sprintf(`UPDATE users SET %v WHERE id=$%v`, strings.Join(fields, ", "), argIdx+1)

	args = append(args, user.ID)

	result, err := u.db.ExecContext(ctx, query, args...)
	if err != nil {
		u.log.Error("failed to execute UPDATE query",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return domain.MakeError(domain.ErrFailedToUpdate, err, "user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		u.log.Error("failed to get rows affected after UPDATE",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return domain.MakeError(domain.ErrFailedToUpdate, err, "user")
	}

	if rowsAffected == 0 {
		return domain.MakeError(domain.ErrFailedToUpdate, domain.ErrNotFound, "user")
	}

	return nil
}

func (u *UserR) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users where id = $1`
	result, err := u.db.ExecContext(ctx, query, userID)
	if err != nil {
		u.log.Error("failed to execute DELETE query",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return domain.MakeError(domain.ErrFailedToDelete, err, "user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		u.log.Error("failed to get rows affected after DELETE",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return domain.MakeError(domain.ErrFailedToDelete, err, "user")
	}

	if rowsAffected == 0 {
		return domain.MakeError(domain.ErrFailedToDelete, domain.ErrNotFound, "user")
	}

	return nil
}
