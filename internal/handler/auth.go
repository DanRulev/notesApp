package handler

import (
	"context"
	"net/http"
	"noteApp/internal/models/dto"
	"noteApp/pkg/logger"
	"noteApp/pkg/valid"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthSI interface {
	SignUp(ctx context.Context, user dto.UserCreate) (uuid.UUID, error)
	SignIn(ctx context.Context, data dto.UserSignIn) (dto.TokenOutput, error)
	Logout(ctx context.Context, tokenID string) error
	ParseToken(ctx context.Context, accessToken string) (uuid.UUID, error)
	RefreshToken(ctx context.Context, tokenID string) (dto.TokenOutput, error)
}

type authH struct {
	service         AuthSI
	refreshTokenTTL time.Duration
	log             *logger.Logger
}

func newAuthHandler(service AuthSI, refreshTokenTTL time.Duration, log *logger.Logger) *authH {
	return &authH{
		service:         service,
		refreshTokenTTL: refreshTokenTTL,
		log:             log,
	}
}

func (h *Handler) InitAuthAPIs(path *gin.RouterGroup) {
	h.log.Info("init auth APIs")
	auth := path.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		auth.GET("/logout", h.logout)
		auth.GET("/refresh", h.refresh)
	}

}

func (h *authH) signUp(c *gin.Context) {
	var user dto.UserCreate

	if err := c.ShouldBindJSON(&user); err != nil {
		h.log.Debug("invalid JSON in sign-up request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := valid.ValidateStruct(user); err != nil {
		h.log.Debug("validation failed during sign-up",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := h.service.SignUp(c.Request.Context(), user)
	if err != nil {
		h.log.Error("sign-up failed",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user signed up successfully",
		zap.String("user_id", userID.String()),
		zap.String("email", user.Email),
		zap.String("client_ip", c.ClientIP()),
	)

	newSuccessResponse(c, http.StatusOK, "id", userID)
}

func (h *authH) signIn(c *gin.Context) {
	var user dto.UserSignIn

	if err := c.ShouldBindJSON(&user); err != nil {
		h.log.Debug("invalid JSON in sign-in request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := valid.ValidateStruct(user); err != nil {
		h.log.Debug("validation failed during sign-in",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.service.SignIn(c.Request.Context(), user)
	if err != nil {
		h.log.Error("sign-in failed",
			zap.Error(err),
			zap.String("email", user.Email),
			zap.String("client_ip", c.ClientIP()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user signed in successfully",
		zap.String("email", user.Email),
		zap.String("client_ip", c.ClientIP()),
	)

	c.SetCookie(refreshToken, token.RefreshToken, int(h.refreshTokenTTL.Seconds()), "/", "", false, true)
	newSuccessResponse(c, http.StatusOK, accessToken, token.AccessToken)
}

func (h *authH) logout(c *gin.Context) {
	tokenID, err := getAccessToken(c)
	if err != nil {
		h.log.Debug("logout failed: no access token",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.service.Logout(c.Request.Context(), tokenID); err != nil {
		h.log.Error("logout failed",
			zap.Error(err),
			zap.String("token_id", tokenID),
			zap.String("client_ip", c.ClientIP()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user logged out",
		zap.String("token_id", tokenID),
		zap.String("client_ip", c.ClientIP()),
	)

	c.SetCookie(refreshToken, "", 0, "/", "", false, true)
	newSuccessResponse(c, http.StatusOK, "message", "logout")
}

func (h *authH) refresh(c *gin.Context) {
	refreshTkn, err := getRefreshToken(c)
	if err != nil {
		h.log.Debug("refresh failed: no refresh token",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.service.RefreshToken(c.Request.Context(), refreshTkn)
	if err != nil {
		h.log.Error("refresh token failed due to internal error",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("token refreshed successfully",
		zap.String("client_ip", c.ClientIP()),
	)

	c.SetCookie(refreshToken, token.RefreshToken, int(h.refreshTokenTTL.Seconds()), "/", "", false, true)
	newSuccessResponse(c, http.StatusOK, accessToken, token.AccessToken)
}
