package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	mock_handler "noteApp/internal/handler/mock"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockAuthHandler(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_handler.MockServiceI)) *Handler {
	t.Helper()

	service := mock_handler.NewMockServiceI(ctrl)

	if setupMock != nil {
		setupMock(service)
	}

	refreshTokenTTL, err := getRefreshTokenTTL()
	require.NoError(t, err)

	return &Handler{
		authH: newAuthHandler(service, refreshTokenTTL, logger.LoggerForTest()),
	}
}

func Test_authH_signUp(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		name                 string
		inputBody            string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "success",
			inputBody: `{"username":"test_user","email":"test_email@gmail.com","password":"qwerty123"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(testUserID, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"` + testUserID.String() + `"}`,
		},
		{
			name:                 "empty body",
			inputBody:            ``,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"EOF"}`,
		},
		{
			name:                 "empty username",
			inputBody:            `{"username":"","email":"test_email@gmail.com","password":"qwerty123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Username, Tag: required, Param: "}`,
		},
		{
			name:                 "username too short",
			inputBody:            `{"username":"ab","email":"test_email@gmail.com","password":"qwerty123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Username, Tag: min, Param: 3"}`,
		},
		{
			name:                 "username too long",
			inputBody:            fmt.Sprintf(`{"username":"%s","email":"test_email@gmail.com","password":"qwerty123"}`, strings.Repeat("a", 256)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Username, Tag: max, Param: 255"}`,
		},
		{
			name:                 "empty email",
			inputBody:            `{"username":"test_user","email":"","password":"qwerty123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: required, Param: "}`,
		},
		{
			name:                 "invalid email",
			inputBody:            `{"username":"test_user","email":"test_emailgmailcom","password":"qwerty123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: email, Param: "}`,
		},
		{
			name:                 "email too long",
			inputBody:            fmt.Sprintf(`{"username":"test_user","email":"%s@gmail.com","password":"qwerty123"}`, strings.Repeat("a", 250)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: max, Param: 255"}`,
		},
		{
			name:                 "empty password",
			inputBody:            `{"username":"test_user","email":"test_email@gmail.com","password":""}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Password, Tag: required, Param: "}`,
		},
		{
			name:                 "password too short",
			inputBody:            `{"username":"test_user","email":"test_email@gmail.com","password":"123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Password, Tag: min, Param: 8"}`,
		},
		{
			name:                 "password too long",
			inputBody:            fmt.Sprintf(`{"username":"test_user","email":"test_email@gmail.com","password":"%s"}`, strings.Repeat("a", 73)),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Password, Tag: max, Param: 72"}`,
		},
		{
			name:                 "invalid image_url",
			inputBody:            `{"username":"test_user","email":"test_email@gmail.com","password":"qwerty123","image_url":"not-a-url"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: ImageURL, Tag: url, Param: "}`,
		},
		{
			name:      "empty image_url",
			inputBody: `{"username":"test_user","email":"test_email@gmail.com","password":"qwerty123","image_url":""}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(testUserID, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"` + testUserID.String() + `"}`,
		},
		{
			name:      "extra fields in body",
			inputBody: `{"username":"test_user","email":"test_email@gmail.com","password":"qwerty123","role":"admin"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(testUserID, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"` + testUserID.String() + `"}`,
		},
		{
			name:      "service error",
			inputBody: `{"username":"test_user","email":"test_email@gmail.com","password":"qwerty123"}`,
			f: func(s *mock_handler.MockServiceI) {
				s.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("service error"))
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

			handler := mockAuthHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/sign-up", handler.signUp)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-up", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func Test_authH_signIn(t *testing.T) {
	t.Parallel()

	token := dto.TokenOutput{
		AccessToken:  "valid-access-token-1234",
		RefreshToken: "valid-refresh-token-1234",
	}

	tests := []struct {
		name                 string
		inputBody            string
		f                    func(*mock_handler.MockServiceI)
		expectedStatusCode   int
		expectedResponseBody string
		expectedCookie       string
	}{
		{
			name:      "success",
			inputBody: `{"email":"test_email@gmail.com", "password":"qwerty123"}`,
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(token, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: fmt.Sprintf(`{"%v":"%v"}`, accessToken, token.AccessToken),
			expectedCookie:       fmt.Sprintf("%v=%v", refreshToken, token.RefreshToken),
		},
		{
			name:                 "empty body",
			inputBody:            ``,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"EOF"}`,
		},
		{
			name:                 "empty email",
			inputBody:            `{"email":"","password":"qwerty123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: required, Param: "}`,
		},
		{
			name:                 "invalid email",
			inputBody:            `{"email":"test_emailgmailcom","password":"qwerty123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Email, Tag: email, Param: "}`,
		},
		{
			name:                 "empty password",
			inputBody:            `{"email":"test_email@gmail.com","password":""}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Password, Tag: required, Param: "}`,
		},
		{
			name:                 "invalid password",
			inputBody:            `{"email":"test_email@gmail.com","password":"123"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"validation failed: Field: Password, Tag: min, Param: 8"}`,
		},
		{
			name:      "service error",
			inputBody: `{"email":"test_email@gmail.com","password":"qwerty123"}`,
			f: func(s *mock_handler.MockServiceI) {
				s.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(dto.TokenOutput{}, errors.New("service error"))
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

			handler := mockAuthHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/sign-in", handler.signIn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-in", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedResponseBody != "" {
				assert.Equal(t, tt.expectedResponseBody, w.Body.String())
			}

			if tt.expectedCookie != "" {
				cookie := w.Result().Cookies()[0]
				assert.Equal(t, refreshToken, cookie.Name)
				assert.Equal(t, token.RefreshToken, cookie.Value)
				assert.Equal(t, int(handler.refreshTokenTTL.Seconds()), cookie.MaxAge)
			}
		})
	}
}

func Test_authH_logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		f                    func(*mock_handler.MockServiceI)
		header               string
		cookie               *http.Cookie
		expectedStatusCode   int
		expectedResponseBody string
		expectedCookie       string
	}{
		{
			name: "success",
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(nil)
			},
			header: "Bearer access-token-1234",
			cookie: &http.Cookie{
				Name:     refreshToken,
				Value:    "refresh-token-1234",
				MaxAge:   5,
				Path:     "/",
				Domain:   "",
				Secure:   false,
				HttpOnly: true,
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"message":"logout"}`,
			expectedCookie:       refreshToken,
		},
		{
			name:                 "failed get refresh token",
			cookie:               &http.Cookie{},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"empty authorization header"}`,
			expectedCookie:       "",
		},
		{
			name: "service error",
			f: func(s *mock_handler.MockServiceI) {
				s.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
			},
			header: "Bearer access-token-1234",
			cookie: &http.Cookie{
				Name:     refreshToken,
				Value:    "refresh-token-1234",
				MaxAge:   5,
				Path:     "/",
				Domain:   "",
				Secure:   false,
				HttpOnly: true,
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
			expectedCookie:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockAuthHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/logout", handler.logout)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/logout", nil)
			req.Header.Add(authHeader, tt.header)
			req.AddCookie(tt.cookie)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, w.Body.String())

			cookies := w.Result().Cookies()
			if tt.expectedCookie == refreshToken {
				require.NotEmpty(t, cookies)

				cookie := cookies[0]
				assert.Equal(t, refreshToken, cookie.Name)
				assert.Equal(t, "", cookie.Value)
				assert.Equal(t, 0, cookie.MaxAge)
			} else {
				require.Empty(t, cookies)
			}
		})
	}
}

func Test_authH_refresh(t *testing.T) {
	t.Parallel()

	token := dto.TokenOutput{
		AccessToken:  "access-token-1234",
		RefreshToken: "refresh-token-1234",
	}

	tests := []struct {
		name                 string
		f                    func(*mock_handler.MockServiceI)
		cookie               *http.Cookie
		expectedStatusCode   int
		expectedResponseBody string
		expectedCookie       string
	}{
		{
			name: "success",
			f: func(msi *mock_handler.MockServiceI) {
				msi.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(token, nil)
			},
			cookie: &http.Cookie{
				Name:     refreshToken,
				Value:    token.RefreshToken,
				MaxAge:   5,
				Path:     "/",
				Domain:   "",
				Secure:   false,
				HttpOnly: true,
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: fmt.Sprintf(`{"%v":"%v"}`, accessToken, token.AccessToken),
			expectedCookie:       refreshToken,
		},
		{
			name:                 "failed to get refresh token",
			cookie:               &http.Cookie{},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"error":"token ID not found in cookie: http: named cookie not present"}`,
			expectedCookie:       "",
		},
		{
			name: "service error",
			f: func(s *mock_handler.MockServiceI) {
				s.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(dto.TokenOutput{}, errors.New("service error"))
			},
			cookie: &http.Cookie{
				Name:     refreshToken,
				Value:    "refresh-token-1234",
				MaxAge:   5,
				Path:     "/",
				Domain:   "",
				Secure:   false,
				HttpOnly: true,
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"service error"}`,
			expectedCookie:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockAuthHandler(t, ctrl, tt.f)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/refresh", handler.refresh)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/refresh", nil)
			req.AddCookie(tt.cookie)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, w.Body.String())

			if tt.expectedCookie == refreshToken {
				cookies := w.Result().Cookies()
				require.NotEmpty(t, cookies)

				cookie := cookies[0]
				assert.Equal(t, tt.cookie.Value, cookie.Value)
				assert.Equal(t, int(handler.refreshTokenTTL.Seconds()), cookie.MaxAge)
				assert.Equal(t, tt.cookie.Name, cookie.Name)
				assert.Equal(t, tt.cookie.Path, cookie.Path)
				assert.Equal(t, tt.cookie.Secure, cookie.Secure)
				assert.Equal(t, tt.cookie.HttpOnly, cookie.HttpOnly)
			}
		})
	}
}
