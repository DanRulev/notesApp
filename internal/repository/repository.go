package repository

import (
	"context"
	"database/sql"
	"noteApp/pkg/logger"
)

type query interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type repository struct {
	*NoteR
	*TokenR
	*UserR
}

func NewRepository(q query, log *logger.Logger) repository {
	return repository{
		NoteR:  NewNoteRepository(q, log),
		TokenR: NewTokenRepository(q, log),
		UserR:  NewUserRepository(q, log),
	}
}
