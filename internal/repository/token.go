package repository

import (
	"context"
	"database/sql"
	"noteApp/internal/models/domain"
	"noteApp/pkg/logger"

	"go.uber.org/zap"
)

type TokenR struct {
	db  query
	log *logger.Logger
}

func NewTokenRepository(db query, log *logger.Logger) *TokenR {
	return &TokenR{
		db:  db,
		log: log,
	}
}

func (t *TokenR) CreateToken(ctx context.Context, token domain.Token) error {
	query := `INSERT INTO tokens (user_id, token_id, expired_at) VALUES ($1, $2, $3)`

	_, err := t.db.ExecContext(ctx, query, token.UserID, token.TokenID, token.ExpiresAt)
	if err != nil {
		t.log.Error("failed to execute INSERT query in CreateToken",
			zap.Error(err),
			zap.String("token_id", token.TokenID),
			zap.String("user_id", token.UserID.String()),
		)
		return domain.MakeError(domain.ErrFailedToCreate, err, "token")
	}

	return nil
}

func (t *TokenR) Token(ctx context.Context, tokenID string) (domain.Token, error) {
	query := `SELECT user_id, token_id, expired_at FROM tokens WHERE token_id=$1`

	row := t.db.QueryRowContext(ctx, query, tokenID)

	var token domain.Token
	if err := row.Scan(
		&token.UserID,
		&token.TokenID,
		&token.ExpiresAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.Token{}, domain.MakeError(domain.ErrReceiving, domain.ErrNotFound, "token")
		}
		t.log.Error("database error in Token query",
			zap.Error(err),
			zap.String("token_id", tokenID),
		)
		return domain.Token{}, domain.MakeError(domain.ErrReceiving, err, "token")
	}

	return token, nil
}

func (t *TokenR) DeleteToken(ctx context.Context, tokenID string) error {
	query := `DELETE FROM tokens WHERE token_id=$1`

	result, err := t.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		t.log.Error("failed to execute DELETE query",
			zap.Error(err),
			zap.String("token_id", tokenID),
		)
		return domain.MakeError(domain.ErrFailedToDelete, err, "token")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.log.Error("failed to get rows affected after DELETE",
			zap.Error(err),
			zap.String("token_id", tokenID),
		)
		return domain.MakeError(domain.ErrFailedToDelete, err, "token")
	}

	if rowsAffected == 0 {
		return domain.MakeError(domain.ErrFailedToDelete, domain.ErrNotFound, "token")
	}

	return nil
}
