package service

import (
	"context"
	"errors"
	"noteApp/internal/models/domain"
	"noteApp/internal/models/dto"
	mock_service "noteApp/internal/service/mock"
	"noteApp/pkg/logger"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockNoteService(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_service.MockRepositoryI)) *NoteS {
	t.Helper()

	repo := mock_service.NewMockRepositoryI(ctrl)
	if setupMock != nil {
		setupMock(repo)
	}

	return NewNoteService(repo, logger.LoggerForTest())
}

func TestNoteS_CreateNote(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		note dto.NoteCreate
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				note: dto.NoteCreate{
					UserID:  uuid.New(),
					Heading: "test_heading",
					Content: "test_content",
					Done:    false,
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty user ID",
			args: args{
				ctx: context.Background(),
				note: dto.NoteCreate{
					UserID:  uuid.Nil,
					Heading: "test_heading",
					Content: "test_content",
					Done:    false,
				},
			},
			f:       nil,
			wantErr: true,
		},
		{
			name: "empty heading",
			args: args{
				ctx: context.Background(),
				note: dto.NoteCreate{
					UserID:  uuid.New(),
					Heading: "",
					Content: "test_content",
					Done:    false,
				},
			},
			f:       nil,
			wantErr: true,
		},
		{
			name: "empty content",
			args: args{
				ctx: context.Background(),
				note: dto.NoteCreate{
					UserID:  uuid.New(),
					Heading: "test_heading",
					Content: "",
					Done:    false,
				},
			},
			f:       nil,
			wantErr: true,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				note: dto.NoteCreate{
					UserID:  uuid.New(),
					Heading: "test_heading",
					Content: "test_content",
					Done:    false,
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockNoteService(t, ctrl, tt.f)

			got, err := repo.CreateNote(tt.args.ctx, tt.args.note)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, got != uuid.Nil)
		})
	}
}

func TestNoteS_Note(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		noteID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.Note{}, nil)
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
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.Note{}, domain.ErrReceiving)
			},
			wantErr: true,
		},
		{
			name: "note not found",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.Note{}, domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "invalid note ID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Note(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.Note{}, domain.ErrReceiving)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockNoteService(t, ctrl, tt.f)

			got, err := repo.Note(tt.args.ctx, tt.args.userID, tt.args.noteID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.IsType(t, got, dto.NoteOutput{})
		})
	}
}

func TestNoteS_Notes(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		p      dto.Paginated
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI)
		want    int
		wantErr bool
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
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]domain.Note{}, 1, nil)
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "success with done status",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 0,
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]domain.Note{}, 1, nil)
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "success with offset",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 2,
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]domain.Note{}, 1, nil)
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "invalid user id",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				p: dto.Paginated{
					Limit:  10,
					Offset: 0,
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, nil)
			},
			want:    0,
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
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, nil)
			},
			want:    0,
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
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, domain.ErrReceiving)
			},
			wantErr: true,
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
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().Notes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, domain.ErrReceiving)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockNoteService(t, ctrl, tt.f)

			got, err := repo.Notes(tt.args.ctx, tt.args.userID, tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, got.Pagination.Total, tt.want)
			assert.Equal(t, got.Pagination.TotalPages, tt.want)
		})
	}
}

func TestNoteS_UpdateNote(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		note dto.NoteUpdate
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil note id",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.Nil,
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},

			wantErr: true,
		},
		{
			name: "nil user id",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.Nil,
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},

			wantErr: true,
		},
		{
			name: "empty heading",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{""}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},
			wantErr: true,
		},
		{
			name: "empty content",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{""}[0],
					Done:    &[]bool{false}[0],
				},
			},
			wantErr: true,
		},
		{
			name: "invalid note id",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "invalid user id",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				note: dto.NoteUpdate{
					ID:      uuid.New(),
					UserID:  uuid.New(),
					Heading: &[]string{"update_heading"}[0],
					Content: &[]string{"update_content"}[0],
					Done:    &[]bool{false}[0],
				},
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockNoteService(t, ctrl, tt.f)

			err := repo.UpdateNote(tt.args.ctx, tt.args.note)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNoteS_DeleteNote(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		noteID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "invalid user id",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrFailedToDelete)
			},
			wantErr: true,
		},
		{
			name: "invalid note id",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				noteID: uuid.New(),
			},
			f: func(mri *mock_service.MockRepositoryI) {
				mri.EXPECT().DeleteNote(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrFailedToDelete)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockNoteService(t, ctrl, tt.f)

			err := repo.DeleteNote(tt.args.ctx, tt.args.userID, tt.args.noteID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
