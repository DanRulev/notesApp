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

type NoteRI interface {
	CreateNote(ctx context.Context, note domain.Note) error
	Note(ctx context.Context, userID, noteID uuid.UUID) (domain.Note, error)
	Notes(ctx context.Context, userID uuid.UUID, p dto.Paginated) ([]domain.Note, int, error)
	UpdateNote(ctx context.Context, note domain.NoteUpdate) error
	DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error
}

type NoteS struct {
	repo NoteRI
	log  *logger.Logger
}

func NewNoteService(repo NoteRI, log *logger.Logger) *NoteS {
	return &NoteS{
		repo: repo,
		log:  log,
	}
}

func (n *NoteS) CreateNote(ctx context.Context, note dto.NoteCreate) (uuid.UUID, error) {
	noteID := uuid.New()
	input := noteCreateDTOtoDomain(note)
	input.ID = noteID

	if err := input.Validate(); err != nil {
		n.log.Debug("note validation failed in service",
			zap.String("user_id", input.UserID.String()),
			zap.String("note_id", input.ID.String()),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	if err := n.repo.CreateNote(ctx, input); err != nil {
		n.log.Error("failed to create note in repository",
			zap.Error(err),
			zap.String("user_id", input.UserID.String()),
			zap.String("note_id", input.ID.String()),
		)
		return uuid.Nil, err
	}

	n.log.Info("note created successfully in service",
		zap.String("user_id", input.UserID.String()),
		zap.String("note_id", input.ID.String()),
	)

	return noteID, nil
}

func (n *NoteS) Note(ctx context.Context, userID, noteID uuid.UUID) (dto.NoteOutput, error) {
	noteDB, err := n.repo.Note(ctx, userID, noteID)
	if err != nil {
		if err == domain.ErrNotFound {
			n.log.Warn("note not found",
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
		} else {
			n.log.Error("failed to get note from repository",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
		}
		return dto.NoteOutput{}, err
	}

	n.log.Debug("note retrieved from repository",
		zap.String("user_id", noteDB.UserID.String()),
		zap.String("note_id", noteDB.ID.String()),
	)

	return noteDomainToDTO(noteDB), nil
}

func (n *NoteS) Notes(ctx context.Context, userID uuid.UUID, p dto.Paginated) (dto.PaginatedResponse, error) {
	notesDB, total, err := n.repo.Notes(ctx, userID, p)
	if err != nil {
		n.log.Error("failed to get notes from repository",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Int("limit", p.Limit),
			zap.Int("offset", p.Offset),
		)
		return dto.PaginatedResponse{}, err
	}

	if total == 0 {
		n.log.Debug("no notes found for user",
			zap.String("user_id", userID.String()),
			zap.Int("limit", p.Limit),
			zap.Int("offset", p.Offset),
		)
		return dto.PaginatedResponse{}, nil
	}

	notes := make([]dto.NoteOutput, len(notesDB))
	for _, v := range notesDB {
		notes = append(notes, noteDomainToDTO(v))
	}

	n.log.Debug("notes fetched and mapped successfully",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(notes)),
		zap.Int("total", total),
		zap.Int("page", p.Limit),
	)

	return dto.MakePaginatedResponse(notes, total, p.Offset, p.Limit), nil
}

func (n *NoteS) UpdateNote(ctx context.Context, note dto.NoteUpdate) error {
	input := noteUpdateDTOtoDomain(note)
	if err := input.Validate(); err != nil {
		return err
	}
	err := n.repo.UpdateNote(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			n.log.Warn("failed to update note",
				zap.String("user_id", input.UserID.String()),
				zap.String("note_id", input.ID.String()),
				zap.Error(err),
			)
		} else {
			n.log.Error("failed to update note in repository",
				zap.Error(err),
				zap.String("user_id", input.UserID.String()),
				zap.String("note_id", input.ID.String()),
			)
		}
		return err
	}

	n.log.Info("note updated successfully",
		zap.String("user_id", input.UserID.String()),
		zap.String("note_id", input.ID.String()),
	)

	return nil
}

func (n *NoteS) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	if err := n.repo.DeleteNote(ctx, userID, noteID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			n.log.Warn("note not found during deletion",
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
		} else {
			n.log.Error("failed to delete note",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
		}
		return err
	}

	n.log.Info("note deleted successfully",
		zap.String("user_id", userID.String()),
		zap.String("note_id", noteID.String()),
	)

	return nil
}

func noteDomainToDTO(note domain.Note) dto.NoteOutput {
	return dto.NoteOutput{
		ID:        note.ID,
		UserID:    note.UserID,
		Heading:   note.Heading,
		Content:   note.Content,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}
}

func noteCreateDTOtoDomain(note dto.NoteCreate) domain.Note {
	return domain.Note{
		ID:      note.ID,
		UserID:  note.UserID,
		Heading: note.Heading,
		Content: note.Content,
	}
}

func noteUpdateDTOtoDomain(note dto.NoteUpdate) domain.NoteUpdate {
	return domain.NoteUpdate{
		ID:      note.ID,
		UserID:  note.UserID,
		Heading: note.Heading,
		Content: note.Content,
	}
}
