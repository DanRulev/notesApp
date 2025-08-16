package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	mock_handler "noteApp/internal/handler/mock"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockNoteHandler(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_handler.MockServiceI)) *Handler {
	t.Helper()

	service := mock_handler.NewMockServiceI(ctrl)

	if setupMock != nil {
		setupMock(service)
	}

	return &Handler{
		noteH: newNoteHandler(service, logger.LoggerForTest()),
	}
}

func Test_noteH_createNote(t *testing.T) {
	t.Parallel()

	noteID := uuid.New()

	tests := []struct {
		name                 string
		inputBody            string
		userID               uuid.UUID
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "success",
			inputBody: `{"heading":"test_heading","content":"test_content","done":false}`,
			userID:    uuid.New(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(noteID, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"` + noteID.String() + `"}`,
		},
		{
			name:                 "missing user_id in context",
			inputBody:            `{"heading":"test_heading","content":"test_content","done":false}`,
			userID:               uuid.Nil,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:                 "empty body",
			userID:               uuid.New(),
			inputBody:            ``,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"EOF"}`,
		},
		{
			name:                 "empty heading",
			inputBody:            `{"heading":"","content":"test_content","done":false}`,
			userID:               uuid.New(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Heading, Tag: required, Param: "}`,
		},
		{
			name:                 "heading too long",
			inputBody:            fmt.Sprintf(`{"heading":"%s","content":"test_content","done":false}`, strings.Repeat("a", 256)),
			userID:               uuid.New(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Heading, Tag: max, Param: 255"}`,
		},
		{
			name:                 "empty content",
			inputBody:            `{"heading":"test_heading","content":"","done":false}`,
			userID:               uuid.New(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Content, Tag: required, Param: "}`,
		},
		{
			name:                 "content too long",
			inputBody:            fmt.Sprintf(`{"heading":"test_heading","content":"%s","done":false}`, strings.Repeat("a", 256)),
			userID:               uuid.New(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Content, Tag: max, Param: 255"}`,
		},
		{
			name:      "service error",
			inputBody: `{"heading":"test_heading","content":"test_content","done":false}`,
			userID:    uuid.New(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(noteID, errors.New("service error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockNoteHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/notes", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, tt.userID.String())
					handler.noteH.createNote(c)
				} else {
					handler.noteH.createNote(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/notes", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_noteH_notes(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	noteID := uuid.New()

	paginatedNotes := dto.PaginatedResponse{
		Data: []dto.NoteOutput{
			{
				ID:        noteID,
				UserID:    userID,
				Heading:   "Test Note",
				Content:   "Content",
				Done:      false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		Pagination: struct {
			Total      int `json:"total"`
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			TotalPages int `json:"total_pages"`
		}{
			Total:      1,
			Page:       0,
			Limit:      10,
			TotalPages: 1,
		},
	}

	bytes, err := json.Marshal(paginatedNotes)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		query                string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:  "success with pagination",
			query: "limit=10&offset=0",
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Notes(gomock.Any(), userID, gomock.Any()).Return(paginatedNotes, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"notes":` + string(bytes) + `}`,
		},
		{
			name:  "success with done=true",
			query: "limit=10&offset=0&done=true",
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Notes(gomock.Any(), userID, gomock.Any()).Return(paginatedNotes, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"notes":` + string(bytes) + `}`,
		},
		{
			name:  "success with done=false",
			query: "limit=10&offset=0&done=false",
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Notes(gomock.Any(), userID, gomock.Any()).Return(paginatedNotes, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"notes":` + string(bytes) + `}`,
		},
		{
			name:                 "invalid limit",
			query:                "limit=-5&offset=0",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Limit, Tag: gte, Param: 10"}`,
		},
		{
			name:                 "invalid offset",
			query:                "limit=10&offset=-5",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Offset, Tag: gte, Param: 0"}`,
		},
		{
			name:                 "missing user_id in context",
			query:                "limit=10&offset=0",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:  "service error",
			query: "limit=10&offset=0",
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return(dto.PaginatedResponse{}, errors.New("service error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockNoteHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/notes", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, userID.String())
					handler.noteH.notes(c)
				} else {
					handler.noteH.notes(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/notes?"+tt.query, nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_noteH_note(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	note := dto.NoteOutput{
		ID:        uuid.New(),
		UserID:    userID,
		Heading:   "Test Note",
		Content:   "Content",
		Done:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bytes, err := json.Marshal(note)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		param                string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:  "success",
			param: note.ID.String(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(note, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"note":` + string(bytes) + `}`,
		},
		{
			name:                 "missing user_id in context",
			param:                note.ID.String(),
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:                 "invalid note_id param",
			param:                "123",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"note_id is not uuid"}`,
		},
		{
			name:  "not found",
			param: note.ID.String(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(dto.NoteOutput{}, domain.ErrNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"not found"}`,
		},
		{
			name:  "service error",
			param: note.ID.String(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(dto.NoteOutput{}, errors.New("service error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockNoteHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/notes/:note_id", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, userID.String())
					handler.noteH.note(c)
				} else {
					handler.noteH.note(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/notes/"+tt.param, nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_noteH_updateNote(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name                 string
		param                string
		inputBody            string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "success",
			param:     uuid.New().String(),
			inputBody: `{"heading":"test_heading","content":"test_content","done":false}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:      "success without heading",
			param:     uuid.New().String(),
			inputBody: `{"content":"test_content","done":false}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:      "success without content",
			param:     uuid.New().String(),
			inputBody: `{"heading":"test_heading","done":false}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:      "success without done",
			param:     uuid.New().String(),
			inputBody: `{"heading":"test_heading","content":"test_content"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:                 "missing user_id in context",
			param:                uuid.New().String(),
			inputBody:            `{"heading":"test_heading","content":"test_content","done":false}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:                 "empty body",
			param:                uuid.New().String(),
			inputBody:            ``,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"EOF"}`,
		},
		{
			name:                 "empty heading",
			param:                uuid.New().String(),
			inputBody:            `{"heading":"","content":"test_content","done":false}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Heading, Tag: min, Param: 1"}`,
		},
		{
			name:                 "heading too long",
			param:                uuid.New().String(),
			inputBody:            fmt.Sprintf(`{"heading":"%s","content":"test_content","done":false}`, strings.Repeat("a", 256)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Heading, Tag: max, Param: 255"}`,
		},
		{
			name:                 "empty content",
			param:                uuid.New().String(),
			inputBody:            `{"heading":"test_heading","content":"","done":false}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Content, Tag: min, Param: 1"}`,
		},
		{
			name:                 "content too long",
			param:                uuid.New().String(),
			inputBody:            fmt.Sprintf(`{"heading":"test_heading","content":"%s","done":false}`, strings.Repeat("a", 256)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Content, Tag: max, Param: 255"}`,
		},
		{
			name:                 "invalid note_id param",
			param:                "123",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"note_id is not uuid"}`,
		},
		{
			name:      "not found",
			param:     uuid.New().String(),
			inputBody: `{"heading":"test_heading","content":"test_content","done":false}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"not found"}`,
		},
		{
			name:      "service error",
			param:     uuid.New().String(),
			inputBody: `{"heading":"test_heading","content":"test_content","done":false}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockNoteHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.PUT("/notes/:note_id", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, userID.String())
					handler.noteH.updateNote(c)
				} else {
					handler.noteH.updateNote(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/notes/"+tt.param, strings.NewReader(tt.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_noteH_deleteNote(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name                 string
		param                string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:  "success",
			param: uuid.New().String(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:                 "missing user_id in context",
			param:                uuid.New().String(),
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:                 "invalid note_id param",
			param:                "123",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"note_id is not uuid"}`,
		},
		{
			name:  "not found",
			param: uuid.New().String(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"not found"}`,
		},
		{
			name:  "service error",
			param: uuid.New().String(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("service error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockNoteHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.DELETE("/notes/:note_id", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, userID.String())
					handler.noteH.deleteNote(c)
				} else {
					handler.noteH.deleteNote(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", "/notes/"+tt.param, nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}
