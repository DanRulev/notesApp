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
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func mockUserHandler(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_handler.MockServiceI)) *Handler {
	t.Helper()

	service := mock_handler.NewMockServiceI(ctrl)

	if setupMock != nil {
		setupMock(service)
	}

	return &Handler{
		userH: newUserHandler(service, logger.LoggerForTest()),
	}
}

func Test_userH_userByID(t *testing.T) {
	t.Parallel()

	userOut := dto.UserOutput{
		ID:        uuid.New(),
		Username:  "test_username",
		Email:     "test_email",
		ImageURL:  "test_image",
		CreatedAt: time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC),
		UpdatedAt: time.Date(2001, time.January, 1, 1, 1, 1, 1, time.UTC),
	}

	bytes, err := json.Marshal(userOut)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		userID               uuid.UUID
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "success",
			userID: userOut.ID,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(userOut, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"user":` + string(bytes) + `}`,
		},
		{
			name:                 "missing user_id in context",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:   "user not found",
			userID: userOut.ID,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(dto.UserOutput{}, domain.ErrNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"not found"}`,
		},
		{
			name:   "service error",
			userID: userOut.ID,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UserByID(gomock.Any(), gomock.Any()).Return(dto.UserOutput{}, errors.New("service error"))
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

			handler := mockUserHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/profile", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, tt.userID.String())
					handler.userByID(c)
				} else {
					handler.userByID(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/profile", nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_userH_updateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		userID               uuid.UUID
		inputBody            string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "success",
			userID:    uuid.New(),
			inputBody: `{"username":"test_username","email":"test@gmail.com","image_url":"http://testimageURL.com"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:      "success without username",
			userID:    uuid.New(),
			inputBody: `{"email":"test@gmail.com","image_url":"http://testimageURL.com"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:      "success without email",
			userID:    uuid.New(),
			inputBody: `{"username":"test_username","image_url":"http://testimageURL.com"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:      "success without image_url",
			userID:    uuid.New(),
			inputBody: `{"username":"test_username","email":"test@gmail.com"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:                 "missing user_id in context",
			inputBody:            `{"username":"test_username","email":"test@gmail.com"}`,
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
			name:                 "empty username",
			userID:               uuid.New(),
			inputBody:            `{"username":"","email":"test@gmail.com","image_url":"http://testimageURL.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Username, Tag: min, Param: 3"}`,
		},
		{
			name:                 "username too long",
			userID:               uuid.New(),
			inputBody:            fmt.Sprintf(`{"username":"%s","email":"test@gmail.com","image_url":"http://testimageURL.com"}`, strings.Repeat("a", 256)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Username, Tag: max, Param: 255"}`,
		},
		{
			name:                 "empty email",
			userID:               uuid.New(),
			inputBody:            `{"username":"test_username","email":"","image_url":"http://testimageURL.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: email, Param: "}`,
		},
		{
			name:                 "email too long",
			userID:               uuid.New(),
			inputBody:            fmt.Sprintf(`{"username":"test_username","email":"%s@gmail.com","image_url":"http://testimageURL.com"}`, strings.Repeat("a", 250)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: max, Param: 255"}`,
		},
		{
			name:                 "empty image_url",
			userID:               uuid.New(),
			inputBody:            `{"username":"test_username","email":"test@gmail.com","image_url":""}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: ImageURL, Tag: url, Param: "}`,
		},
		{
			name:                 "invalid image_url",
			userID:               uuid.New(),
			inputBody:            `{"username":"test_username","email":"test@gmail.com","image_url":"httptestimageURL.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: ImageURL, Tag: url, Param: "}`,
		},
		{
			name:      "service error",
			userID:    uuid.New(),
			inputBody: `{"username":"test_username","email":"test@gmail.com"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
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

			handler := mockUserHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.PUT("/profile", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, tt.userID.String())
					handler.updateUser(c)
				} else {
					handler.updateUser(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/profile", strings.NewReader(tt.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_userH_updateUserPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		userID               uuid.UUID
		inputBody            string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "success",
			userID:    uuid.New(),
			inputBody: `{"old_password":"old_pass","new_password":"new_pass"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:                 "missing user_id in context",
			inputBody:            `{"old_password":"old_pass","new_password":"new_pass"}`,
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
			name:                 "empty old password",
			userID:               uuid.New(),
			inputBody:            `{"old_password":"","new_password":"new_pass"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: OldPassword, Tag: required, Param: "}`,
		},
		{
			name:                 "short old password",
			userID:               uuid.New(),
			inputBody:            `{"old_password":"ret","new_password":"new_pass"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: OldPassword, Tag: min, Param: 8"}`,
		},
		{
			name:                 "old password too long",
			userID:               uuid.New(),
			inputBody:            fmt.Sprintf(`{"old_password":"%s","new_password":"new_pass"}`, strings.Repeat("a", 73)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: OldPassword, Tag: max, Param: 72"}`,
		},
		{
			name:                 "empty new password",
			userID:               uuid.New(),
			inputBody:            `{"old_password":"old_password","new_password":""}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: NewPassword, Tag: required, Param: "}`,
		},
		{
			name:                 "short new password",
			userID:               uuid.New(),
			inputBody:            `{"old_password":"old_password","new_password":"ret"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: NewPassword, Tag: min, Param: 8"}`,
		},
		{
			name:                 "new password too long",
			userID:               uuid.New(),
			inputBody:            fmt.Sprintf(`{"old_password":"old_password","new_password":"%s"}`, strings.Repeat("a", 73)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: NewPassword, Tag: max, Param: 72"}`,
		},
		{
			name:      "incorrect password",
			userID:    uuid.New(),
			inputBody: `{"old_password":"old_pass","new_password":"incorrect"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).Return(domain.ErrIncorrectPassword)
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"incorrect password"}`,
		},
		{
			name:      "service error",
			userID:    uuid.New(),
			inputBody: `{"old_password":"old_pass","new_password":"incorrect"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
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

			handler := mockUserHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.PUT("/profile/pass", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, tt.userID.String())
					handler.updateUserPass(c)
				} else {
					handler.updateUserPass(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/profile/pass", strings.NewReader(tt.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_userH_deleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		userID               uuid.UUID
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "success",
			userID: uuid.New(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"data":"ok"}`,
		},
		{
			name:                 "missing user_id in context",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"user id not found in context"}`,
		},
		{
			name:   "service error",
			userID: uuid.New(),
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
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

			handler := mockUserHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.DELETE("/profile", func(c *gin.Context) {
				if tt.name != "missing user_id in context" {
					c.Set(userIDKey, tt.userID.String())
					handler.deleteUser(c)
				} else {
					handler.deleteUser(c)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", "/profile", nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}
