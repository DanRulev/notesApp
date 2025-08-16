package handler

import (
	"net/http"
	"noteApp/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceI interface {
	AuthSI
	NoteSI
	UserSI
}

type Handler struct {
	*authH
	*noteH
	*userH
	log *logger.Logger
}

func NewHandler(service ServiceI, log *logger.Logger, refreshTokenTTL time.Duration) *Handler {
	return &Handler{
		authH: newAuthHandler(service, refreshTokenTTL, log),
		noteH: newNoteHandler(service, log),
		userH: newUserHandler(service, log),
		log:   log,
	}
}

func (h *Handler) Init() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		h.requestID(),
		h.logging(),
		gin.Recovery(),
	)

	h.initAPIs(router)
	return router
}

func (h *Handler) initAPIs(router *gin.Engine) {
	router.GET("/ping", h.ping)
	api := router.Group("/api")
	{
		api.GET("/home", h.home)
		h.InitAuthAPIs(api)
		h.InitNoteAPIs(api)
		h.InitUserAPIs(api)
	}
}

func (h *Handler) home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (h *Handler) ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
