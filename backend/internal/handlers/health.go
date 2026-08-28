package handlers

import (
	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db      *database.DB
	appEnv  string
	version string
}

func NewHealthHandler(db *database.DB, appEnv, version string) *HealthHandler {
	return &HealthHandler{db: db, appEnv: appEnv, version: version}
}

func (h *HealthHandler) Status(c *gin.Context) {
	data := gin.H{
		"status":      "ok",
		"database":    "up",
		"environment": h.appEnv,
		"version":     h.version,
	}

	if err := h.db.Ping(); err != nil {
		data["status"] = "degraded"
		data["database"] = "down"
		c.JSON(503, httpx.Response{
			Success: false,
			Data:    data,
			Error:   apperr.Wrap(err, 503, apperr.CodeDatabase, "اتصال دیتابیس برقرار نیست."),
		})
		return
	}

	httpx.OK(c, 200, data)
}
