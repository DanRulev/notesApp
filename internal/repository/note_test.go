package repository

import (
	"context"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNoteR_CreateNote(t *testing.T) {
	type args struct {
		ctx  context.Context
		note domain.Note
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
				context.Background(),
				domain.Note{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: "test_heading",
					Content: "test_content",
				},
			},
			wantErr: false,
		},
		{
			name: "without user",
			args: args{
				context.Background(),
				domain.Note{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: "test_heading",
					Content: "test_content",
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty heading",
			args: args{
				context.Background(),
				domain.Note{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: "",
					Content: "test_content",
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToCreate.Error(),
		},
		{
			name: "empty content",
			args: args{
				context.Background(),
				domain.Note{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: "test_heading",
					Content: "",
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
			t.Cleanup(func() { _ = tx.Rollback() })

			repoNote := NewNoteRepository(tx, logger.LoggerForTest())

			if tt.name != "without user" {
				repoUser := NewUserRepository(tx, logger.LoggerForTest())
				err := repoUser.CreateUser(context.Background(), domain.User{
					ID:       tt.args.note.UserID,
					Username: "test",
					Email:    "test",
					Password: "test",
					ImageURL: "test",
				})
				require.NoError(t, err)
			}

			err = repoNote.CreateNote(tt.args.ctx, tt.args.note)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}
			require.NoError(t, err)

			got, err := repoNote.Note(context.Background(), tt.args.note.UserID, tt.args.note.ID)
			require.NoError(t, err)
			assert.Equal(t, got.Heading, tt.args.note.Heading)
			assert.Equal(t, got.Content, tt.args.note.Content)
			require.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
			require.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
		})
	}
}

func TestNoteR_Note(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		noteID uuid.UUID
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
				noteID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "wrong note ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			wantErr:    true,
			wantErrMsg: domain.ErrReceiving.Error(),
		},
		{
			name: "wrong user ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
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

			repoUser := NewUserRepository(tx, logger.LoggerForTest())
			err = repoUser.CreateUser(context.Background(), domain.User{
				ID:       tt.args.userID,
				Username: "test",
				Email:    "test",
				Password: "test",
				ImageURL: "test",
			})
			require.NoError(t, err)

			repo := NewNoteRepository(tx, logger.LoggerForTest())

			testNote := domain.Note{
				ID:      tt.args.noteID,
				UserID:  tt.args.userID,
				Heading: "test_heading",
				Content: "test_content",
			}
			err = repo.CreateNote(context.Background(), testNote)
			require.NoError(t, err)

			userID, noteID := tt.args.userID, tt.args.noteID
			switch tt.name {
			case "wrong note ID":
				noteID = uuid.New()
			case "wrong user ID":
				userID = uuid.New()
			default:
			}

			got, err := repo.Note(tt.args.ctx, userID, noteID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, testNote.ID, got.ID)
			assert.Equal(t, testNote.UserID, got.UserID)
			assert.Equal(t, testNote.Heading, got.Heading)
			assert.Equal(t, testNote.Content, got.Content)
			require.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
			require.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)

		})
	}
}

func TestNoteR_Notes(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		p      dto.Paginated
	}
	tests := []struct {
		name       string
		args       args
		want       []domain.Note
		want1      int
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 0,
				},
			},
			want1:   5,
			wantErr: false,
		},
		{
			name: "success with paginated limit",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  3,
					Offset: 0,
				},
			},
			want1:   3,
			wantErr: false,
		},
		{
			name: "success with paginated offset",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 2,
				},
			},
			want1:   3,
			wantErr: false,
		},
		{
			name: "no notes for user",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 0,
				},
			},
			want1:   0,
			wantErr: false,
		},
		{
			name: "offset exceeds total",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 10,
				},
			},
			want:    nil,
			want1:   0,
			wantErr: false,
		},
		{
			name: "negative limit",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  -1,
					Offset: 0,
				},
			},
			want:       nil,
			want1:      0,
			wantErr:    true,
			wantErrMsg: domain.ErrReceiving.Error(),
		},
		{
			name: "negative offset",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: -1,
				},
			},
			want:       nil,
			want1:      0,
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

			repo := NewNoteRepository(tx, logger.LoggerForTest())

			repoUser := NewUserRepository(tx, logger.LoggerForTest())
			err = repoUser.CreateUser(context.Background(), domain.User{
				ID:       tt.args.userID,
				Username: "test",
				Email:    "test",
				Password: "test",
				ImageURL: "test",
			})

			require.NoError(t, err)

			if tt.name != "no notes for user" {
				for range 5 {
					note := domain.Note{
						ID:      uuid.New(),
						UserID:  tt.args.userID,
						Heading: "test_heading",
						Content: "test_content",
					}

					err := repo.CreateNote(context.Background(), note)
					require.NoError(t, err)
				}
			}

			got, got1, err := repo.Notes(tt.args.ctx, tt.args.userID, tt.args.p)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want1, len(got))
			require.LessOrEqual(t, tt.want1, got1)
			for _, note := range got {
				assert.Equal(t, note.UserID, tt.args.userID)
				assert.Equal(t, note.Heading, "test_heading")
				assert.Equal(t, note.Content, "test_content")
				require.WithinDuration(t, time.Now(), note.CreatedAt, time.Second)
				require.WithinDuration(t, time.Now(), note.UpdatedAt, time.Second)
			}
		})
	}
}

func TestNoteR_UpdateNote(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		note domain.NoteUpdate
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
				note: domain.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_test_heading"}[0],
					Content: &[]string{"update_test_content"}[0],
				},
			},
			wantErr: false,
		},
		{
			name: "nil fields",
			args: args{
				ctx:  context.Background(),
				note: domain.NoteUpdate{},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrNoFieldsToUpdate.Error(),
		},
		{
			name: "empty heading",
			args: args{
				ctx: context.Background(),
				note: domain.NoteUpdate{
					Heading: &[]string{""}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
		{
			name: "empty content",
			args: args{
				ctx: context.Background(),
				note: domain.NoteUpdate{
					Content: &[]string{""}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToUpdate.Error(),
		},
		{
			name: "invalid note ID",
			args: args{
				ctx: context.Background(),
				note: domain.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"test_heading"}[0],
					Content: &[]string{"test_content"}[0],
				},
			},
			wantErr:    true,
			wantErrMsg: domain.ErrNotFound.Error(),
		},
		{
			name: "invalid user ID",
			args: args{
				ctx: context.Background(),
				note: domain.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"test_heading"}[0],
					Content: &[]string{"test_content"}[0],
				},
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

			repo := NewNoteRepository(tx, logger.LoggerForTest())

			userID, noteID := tt.args.note.UserID, tt.args.note.ID
			switch tt.name {
			case "invalid note ID":
				noteID = uuid.New()
			case "invalid user ID":
				userID = uuid.New()
			default:
			}

			repoUser := NewUserRepository(tx, logger.LoggerForTest())
			err = repoUser.CreateUser(context.Background(), domain.User{
				ID:       userID,
				Username: "test",
				Email:    "test",
				Password: "test",
				ImageURL: "test",
			})

			require.NoError(t, err)

			testNote := domain.Note{
				ID:      noteID,
				UserID:  userID,
				Heading: "test_heading",
				Content: "test_content",
			}
			err = repo.CreateNote(context.Background(), testNote)
			require.NoError(t, err)

			err = repo.UpdateNote(tt.args.ctx, tt.args.note)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			noteDB, err := repo.Note(context.Background(), tt.args.note.UserID, tt.args.note.ID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.args.note.Heading != nil {
				assert.Equal(t, noteDB.Heading, *tt.args.note.Heading)
			} else {
				assert.Equal(t, noteDB.Heading, testNote.Heading)
			}

			if tt.args.note.Content != nil {
				assert.Equal(t, noteDB.Content, *tt.args.note.Content)
			} else {
				assert.Equal(t, noteDB.Content, testNote.Content)
			}

			require.WithinDuration(t, time.Now(), noteDB.CreatedAt, time.Second)

			require.WithinDuration(t, time.Now(), noteDB.UpdatedAt, time.Second)
			require.Greater(t, noteDB.UpdatedAt.Unix(), testNote.CreatedAt.Unix())
		})
	}
}

func TestNoteR_DeleteNote(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		noteID uuid.UUID
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
				noteID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToDelete.Error(),
		},
		{
			name: "invalid note ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			wantErr:    true,
			wantErrMsg: domain.ErrFailedToDelete.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := globalTestDB.BeginTxx(context.Background(), nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = tx.Rollback() })

			repo := NewNoteRepository(tx, logger.LoggerForTest())

			userID, noteID := tt.args.userID, tt.args.noteID
			switch tt.name {
			case "invalid note ID":
				noteID = uuid.New()
			case "invalid user ID":
				userID = uuid.New()
			default:
			}

			repoUser := NewUserRepository(tx, logger.LoggerForTest())
			err = repoUser.CreateUser(context.Background(), domain.User{
				ID:       userID,
				Username: "test",
				Email:    "test",
				Password: "test",
				ImageURL: "test",
			})

			require.NoError(t, err)

			testNote := domain.Note{
				ID:      noteID,
				UserID:  userID,
				Heading: "test_heading",
				Content: "test_content",
			}
			err = repo.CreateNote(context.Background(), testNote)
			require.NoError(t, err)

			err = repo.DeleteNote(tt.args.ctx, tt.args.userID, tt.args.noteID)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)

			_, err = repo.Note(context.Background(), tt.args.userID, tt.args.noteID)
			require.Error(t, err)
		})
	}
}
