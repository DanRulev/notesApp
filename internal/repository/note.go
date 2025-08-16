package repository

import (
	"context"
	"database/sql"
	"fmt"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"noteApp/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NoteR struct {
	db  query
	log *logger.Logger
}

func NewNoteRepository(db query, log *logger.Logger) *NoteR {
	return &NoteR{
		db:  db,
		log: log,
	}
}

func (n *NoteR) CreateNote(ctx context.Context, note domain.Note) error {
	query := `INSERT INTO notes (id, user_id, heading, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := n.db.ExecContext(ctx, query, note.ID, note.UserID, note.Heading, note.Content, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		n.log.Error("failed to execute INSERT query in CreateNote",
			zap.Error(err),
			zap.String("note_id", note.ID.String()),
			zap.String("user_id", note.UserID.String()),
		)
		return domain.MakeError(domain.ErrFailedToCreate, err, "note")
	}

	return nil
}

func (n *NoteR) Note(ctx context.Context, userID, noteID uuid.UUID) (domain.Note, error) {
	query := `
		SELECT
			id,
			user_id,
			heading,
			content,
			created_at,
			updated_at
		FROM notes
		WHERE id=$1 AND user_id=$2`

	var note domain.Note
	err := n.db.QueryRowContext(ctx, query, noteID, userID).Scan(
		&note.ID,
		&note.UserID,
		&note.Heading,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Note{}, domain.MakeError(domain.ErrReceiving, domain.ErrNotFound, "note")
		}
		n.log.Error("database error in Note query",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("note_id", noteID.String()),
		)
		return domain.Note{}, domain.MakeError(domain.ErrReceiving, err, "note")
	}

	return note, nil
}

func (n *NoteR) Notes(ctx context.Context, userID uuid.UUID, p dto.Paginated) ([]domain.Note, int, error) {
	var total int
	query := "SELECT COUNT(*) FROM notes WHERE user_id=$1"
	err := n.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return nil, 0, domain.MakeError(domain.ErrReceiving, err, "notes")
	}

	if total == 0 || total <= p.Offset {
		return nil, 0, nil
	}

	query = `
        SELECT id, user_id, heading, content, created_at, updated_at
        FROM notes
        WHERE user_id=$1
		ORDER BY created_at ASC
        LIMIT $2 OFFSET $3`

	rows, err := n.db.QueryContext(ctx, query, userID, p.Limit, p.Offset)
	if err != nil {
		n.log.Error("failed to execute SELECT query in Notes",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, 0, domain.MakeError(domain.ErrReceiving, err, "notes")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Логгируй, если нужно
			n.log.Error("failed to close rows", zap.Error(closeErr))
		}
	}()

	if p.Offset+p.Limit > total {
		p.Limit = total - p.Offset
	}

	notes := make([]domain.Note, 0, p.Limit)
	for rows.Next() {
		var note domain.Note
		err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.Heading,
			&note.Content,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, 0, domain.MakeError(domain.ErrReceiving, err, "notes")
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		n.log.Error("error during row iteration",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, 0, domain.MakeError(domain.ErrReceiving, err, "notes")
	}

	return notes, total, nil
}

func (n *NoteR) UpdateNote(ctx context.Context, note domain.NoteUpdate) error {
	var (
		fields []string
		args   []interface{}
		argIdx int
	)

	utils.AddFieldsToQuery("heading", note.Heading, &fields, &args, &argIdx)
	utils.AddFieldsToQuery("content", note.Content, &fields, &args, &argIdx)

	if len(fields) == 0 {
		return domain.MakeError(domain.ErrFailedToUpdate, domain.ErrNoFieldsToUpdate, "note")
	}

	fields = append(fields, "updated_at=NOW()")

	query := fmt.Sprintf(`UPDATE notes SET %v WHERE id=$%v AND user_id=$%v`, strings.Join(fields, ", "), argIdx+1, argIdx+2)

	args = append(args, note.ID, note.UserID)

	result, err := n.db.ExecContext(ctx, query, args...)
	if err != nil {
		n.log.Error("failed to execute UPDATE query",
			zap.Error(err),
			zap.String("note_id", note.ID.String()),
			zap.String("user_id", note.UserID.String()),
		)
		return domain.MakeError(domain.ErrFailedToUpdate, err, "note")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		n.log.Error("failed to get rows affected after UPDATE",
			zap.Error(err),
			zap.String("note_id", note.ID.String()),
			zap.String("user_id", note.UserID.String()),
		)
		return domain.MakeError(domain.ErrFailedToUpdate, err, "note")
	}

	if rowsAffected == 0 {
		return domain.MakeError(domain.ErrFailedToUpdate, domain.ErrNotFound, "note")
	}

	return nil
}

func (n *NoteR) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	query := `DELETE FROM notes WHERE id=$1 AND user_id=$2`

	result, err := n.db.ExecContext(ctx, query, noteID, userID)
	if err != nil {
		n.log.Error("failed to execute DELETE query",
			zap.Error(err),
			zap.String("note_id", noteID.String()),
			zap.String("user_id", userID.String()),
		)
		return domain.MakeError(domain.ErrFailedToDelete, err, "note")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		n.log.Error("failed to get rows affected after DELETE",
			zap.Error(err),
			zap.String("note_id", noteID.String()),
			zap.String("user_id", userID.String()),
		)
		return domain.MakeError(domain.ErrFailedToDelete, err, "note")
	}

	if rowsAffected == 0 {
		return domain.MakeError(domain.ErrFailedToDelete, sql.ErrNoRows, "note")
	}

	return nil
}
