package handler

import (
	"context"
	"errors"
	"net/http"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"noteApp/pkg/valid"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NoteSI interface {
	CreateNote(ctx context.Context, note dto.NoteCreate) (uuid.UUID, error)
	Note(ctx context.Context, userID, nodeID uuid.UUID) (dto.NoteOutput, error)
	Notes(ctx context.Context, userID uuid.UUID, p dto.Paginated) (dto.PaginatedResponse, error)
	UpdateNote(ctx context.Context, note dto.NoteUpdate) error
	DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error
}

type noteH struct {
	service NoteSI
	log     *logger.Logger
}

func newNoteHandler(service NoteSI, log *logger.Logger) *noteH {
	return &noteH{
		service: service,
		log:     log,
	}
}

func (h *Handler) InitNoteAPIs(api *gin.RouterGroup) {
	h.log.Info("init notes APIs")
	note := api.Group("/notes", h.authMiddleware)
	{
		note.POST("/", h.createNote)
		note.GET("/", h.notes)
		note.GET("/:note_id", h.note)
		note.PUT("/:note_id", h.updateNote)
		note.DELETE("/:note_id", h.deleteNote)
	}
}

func (n *noteH) createNote(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		n.log.Debug("unauthorized access attempt",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var note dto.NoteCreate

	if err := c.ShouldBindJSON(&note); err != nil {
		n.log.Debug("invalid JSON in create note request",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	note.UserID = userID

	if err := valid.ValidateStruct(note); err != nil {
		n.log.Debug("validation failed for create note",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := n.service.CreateNote(c.Request.Context(), note)
	if err != nil {
		n.log.Error("create note failed",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	n.log.Info("note created successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", userID.String()),
		zap.String("note_id", id.String()),
	)

	newSuccessResponse(c, http.StatusOK, "id", id)
}

func (n *noteH) notes(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		n.log.Debug("unauthorized access attempt",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var p dto.Paginated
	if err := c.ShouldBindQuery(&p); err != nil {
		n.log.Debug("invalid query in notes request",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := valid.ValidateStruct(p); err != nil {
		n.log.Debug("validation failed for notes request",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	notes, err := n.service.Notes(c.Request.Context(), userID, p)
	if err != nil {
		n.log.Error("get notes failed",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	n.log.Info("notes retrieved successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", userID.String()),
		zap.Int("page", notes.Pagination.Page),
		zap.Int("count", notes.Pagination.Limit),
		zap.Int("total", notes.Pagination.Total),
	)

	newSuccessResponse(c, http.StatusOK, "notes", notes)
}

func (n *noteH) note(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		n.log.Debug("unauthorized access attempt",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	noteID, err := getParamUUID(c, "note_id")
	if err != nil {
		n.log.Debug("invalid note_id in URL",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("param_value", c.Param("note_id")),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	note, err := n.service.Note(c.Request.Context(), userID, noteID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			n.log.Warn("note not found",
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		n.log.Error("failed to get note",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("note_id", noteID.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	n.log.Info("note retrieved successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", userID.String()),
		zap.String("note_id", note.ID.String()),
	)

	newSuccessResponse(c, http.StatusOK, "note", note)
}

func (n *noteH) updateNote(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		n.log.Debug("unauthorized access attempt",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	noteID, err := getParamUUID(c, "note_id")
	if err != nil {
		n.log.Debug("invalid note_id in URL",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("param_value", c.Param("note_id")),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var note dto.NoteUpdate
	if err := c.ShouldBindJSON(&note); err != nil {
		n.log.Debug("invalid JSON in update note request",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	note.UserID = userID
	note.ID = noteID

	if err := valid.ValidateStruct(note); err != nil {
		n.log.Debug("validation failed for update note",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("note_id", noteID.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := n.service.UpdateNote(c.Request.Context(), note); err != nil {
		if err == domain.ErrNotFound {
			n.log.Warn("note not found",
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		n.log.Error("failed to update note",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("note_id", noteID.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	n.log.Info("note update successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", userID.String()),
		zap.String("note_id", note.ID.String()),
	)

	newSuccessResponse(c, http.StatusOK, "data", "ok")
}

func (n *noteH) deleteNote(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		n.log.Debug("unauthorized access attempt in deleteNote",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	noteID, err := getParamUUID(c, "note_id")
	if err != nil {
		n.log.Debug("invalid note_id in URL",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("param_value", c.Param("note_id")),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := n.service.DeleteNote(c.Request.Context(), userID, noteID); err != nil {
		if err == domain.ErrNotFound {
			n.log.Warn("note not found",
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_id", userID.String()),
				zap.String("note_id", noteID.String()),
			)
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		n.log.Error("failed to delete note",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID.String()),
			zap.String("note_id", noteID.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	n.log.Info("note delete successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", userID.String()),
		zap.String("note_id", noteID.String()),
	)

	newSuccessResponse(c, http.StatusOK, "data", "ok")
}
