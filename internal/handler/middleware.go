package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	userIDKey      = "user_id"
	roleKey        = "role"
	adminKey       = "admin"
	userKey        = "user"
	refreshToken   = "refresh_token"
	accessToken    = "access_token"
	authHeader     = "Authorization"
	requestHeader  = "X-Request-ID"
	requestContext = "request_id"
)

func (h *Handler) logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		status := c.Writer.Status()
		msg := ""

		var logFunc func(msg string, fields ...zap.Field)
		if status >= 500 {
			logFunc = h.log.Error
			msg = "HTTP Request Failed"
		} else if status >= 400 {
			logFunc = h.log.Warn
			msg = "HTTP Request Bad"
		} else {
			logFunc = h.log.Info
			msg = "HTTP Request"
		}

		logFunc(msg,
			zap.String("request_id", c.GetString("request_id")),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("query", c.Request.URL.RawQuery),
		)
	}
}

func (h *Handler) requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(requestContext, requestID)
		c.Header(requestHeader, requestID)
		c.Next()
	}
}

func (h *Handler) authMiddleware(c *gin.Context) {
	_, err := getRefreshToken(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/api/auth/login")
		c.Abort()
		return
	}

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.Redirect(http.StatusUnauthorized, "/api/auth/login")
		c.Abort()
		return
	}

	userID, err := h.authH.service.ParseToken(c.Request.Context(), accessToken)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/api/auth/login")
		c.Abort()
		return
	}

	c.Set(userIDKey, userID)

	c.Next()
}

func getAccessToken(c *gin.Context) (string, error) {
	token := c.GetHeader(authHeader)
	if token == "" {
		return "", errors.New("empty authorization header")
	}

	tokenPaths := strings.Split(token, " ")
	if len(tokenPaths) != 2 || tokenPaths[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return tokenPaths[1], nil
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	id, exists := c.Get(userIDKey)
	if !exists {
		return uuid.Nil, fmt.Errorf("user id not found in context")
	}

	t, ok := id.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user id format")
	}

	userID, err := uuid.Parse(t)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user id is not uuid")
	}

	if userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid user id")
	}

	return userID, nil
}

func getRefreshToken(c *gin.Context) (string, error) {
	tokenID, err := c.Cookie(refreshToken)
	if err != nil || tokenID == "" {
		return "", fmt.Errorf("token ID not found in cookie: %v", err)
	}

	return tokenID, nil
}

func getParamUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id := c.Param(param)
	res, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%v is not uuid", param)
	}
	return res, nil
}
