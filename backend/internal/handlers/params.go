package handlers

import (
	"strconv"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/repository"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type FinancialHandlers struct {
	services *services.Services
	audits   repository.AuditRepository
	backups  *services.BackupService
}

func NewFinancialHandlers(svcs *services.Services, backups *services.BackupService) *FinancialHandlers {
	return &FinancialHandlers{services: svcs, backups: backups}
}

func parseTimeParam(c *gin.Context, name string) (time.Time, bool) {
	raw := c.Query(name)
	if raw == "" {
		return time.Time{}, true
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	httpx.HandleError(c, apperr.Validation("پارامتر %s فرمت تاریخ معتبر ندارد.", name))
	return time.Time{}, false
}

func optionalID(c *gin.Context, name string) (*int64, bool) {
	raw := c.Query(name)
	if raw == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		httpx.HandleError(c, apperr.Validation("پارامتر %s باید شناسه عددی مثبت باشد.", name))
		return nil, false
	}
	return &id, true
}

func pathID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.Validation("شناسه نامعتبر است.")
	}
	return id, nil
}
