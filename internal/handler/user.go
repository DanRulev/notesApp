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

type UserSI interface {
	UserByID(ctx context.Context, id uuid.UUID) (dto.UserOutput, error)
	UpdateUser(ctx context.Context, user dto.UserUpdate) error
	UpdateUserPassword(ctx context.Context, updPass dto.UserUpdPassword) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type userH struct {
	service UserSI
	log     *logger.Logger
}

func newUserHandler(service UserSI, log *logger.Logger) *userH {
	return &userH{
		service: service,
		log:     log,
	}
}

func (h *Handler) InitUserAPIs(path *gin.RouterGroup) {
	h.log.Info("init user APIs")
	user := path.Group("/profile", h.authMiddleware)
	{
		user.GET("/", h.userH.userByID)
		user.PUT("/", h.userH.updateUser)
		user.PUT("/pass", h.userH.updateUserPass)
		user.DELETE("/", h.userH.deleteUser)
	}
}

func (h *userH) userByID(c *gin.Context) {
	id, err := getUserID(c)
	if err != nil {
		h.log.Debug("unauthorized access attempt in getUserByID",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := h.service.UserByID(c, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.log.Warn("user not found",
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_id", id.String()),
			)
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		h.log.Error("failed to get user by ID",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", id.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user profile retrieved successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", id.String()),
		zap.String("email", user.Email),
	)

	newSuccessResponse(c, http.StatusOK, "user", user)
}

func (h *userH) updateUser(c *gin.Context) {
	id, err := getUserID(c)
	if err != nil {
		h.log.Debug("unauthorized access attempt in updateUser",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var user dto.UserUpdate
	if err := c.ShouldBindJSON(&user); err != nil {
		h.log.Debug("invalid JSON in updateUser request",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", id.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user.ID = id

	if err := valid.ValidateStruct(user); err != nil {
		h.log.Debug("validation failed for updateUser",
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", id.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateUser(c, user); err != nil {
		h.log.Error("failed to update user",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", id.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user updated successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", id.String()),
	)

	newSuccessResponse(c, http.StatusOK, "data", "ok")
}

func (h *userH) updateUserPass(c *gin.Context) {
	id, err := getUserID(c)
	if err != nil {
		h.log.Debug("unauthorized access attempt in updateUserPass",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var updPass dto.UserUpdPassword
	if err := c.ShouldBindJSON(&updPass); err != nil {
		h.log.Debug("invalid JSON in updateUserPass request",
			zap.String("user_id", id.String()),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updPass.UserID = id

	if err := valid.ValidateStruct(updPass); err != nil {
		h.log.Debug("validation failed for updateUserPass",
			zap.String("user_id", id.String()),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateUserPassword(c.Request.Context(), updPass); err != nil {
		if errors.Is(err, domain.ErrIncorrectPassword) {
			h.log.Warn("incorrect old password during password change",
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_id", id.String()),
			)
			newErrorResponse(c, http.StatusUnauthorized, err.Error())
			return
		}

		h.log.Error("failed to update user password",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", id.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user password updated successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", id.String()),
	)

	newSuccessResponse(c, http.StatusOK, "data", "ok")
}

func (h *userH) deleteUser(c *gin.Context) {
	id, err := getUserID(c)
	if err != nil {
		h.log.Debug("unauthorized access attempt in deleteUser",
			zap.String("client_ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.service.DeleteUser(c, id); err != nil {
		h.log.Error("failed to delete user",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", id.String()),
		)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("user deleted successfully",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_id", id.String()),
	)

	newSuccessResponse(c, http.StatusOK, "data", "ok")
}
