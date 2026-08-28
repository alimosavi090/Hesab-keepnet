package handlers

import (
	"net/http"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

const backupIntervalHours = 24

// ─── Notes (data-attached + daily journal) ──────────────────────

type noteRequest struct {
	EntityType string   `json:"entity_type" binding:"required"`
	EntityID   *int64   `json:"entity_id"`
	Body       string   `json:"body" binding:"required"`
	Tags       []string `json:"tags"`
	Pinned     bool     `json:"pinned"`
}

func parseNoteEntityType(raw string) (models.NoteEntityType, bool) {
	t := models.NoteEntityType(raw)
	switch t {
	case models.NoteEntityRepresentative,
		models.NoteEntitySale,
		models.NoteEntityBankAccount,
		models.NoteEntityJournal:
		return t, true
	default:
		return "", false
	}
}

func (h *FinancialHandlers) ListNotes(c *gin.Context) {
	pq, err := queryPage(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	filter := services.NoteFilter{PageQuery: pq}

	if raw := c.Query("entity_type"); raw != "" {
		entityType, ok := parseNoteEntityType(raw)
		if !ok {
			httpx.HandleError(c, apperr.Validation("نوع پیوست یادداشت نامعتبر است."))
			return
		}
		filter.EntityType = entityType
		if entityType == models.NoteEntityJournal {
			filter.JournalOnly = true
		}
	}
	filter.EntityID, _ = optionalID(c, "entity_id")
	if raw := c.Query("pinned"); raw != "" {
		pinned := raw == "true" || raw == "1"
		filter.Pinned = &pinned
	}
	filter.Tag = c.Query("tag")
	filter.Query = c.Query("q")

	items, total, err := h.services.Notes.List(c.Request.Context(), filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	okPaged(c, items, total, pq)
}

func (h *FinancialHandlers) CreateNote(c *gin.Context) {
	var in noteRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی یادداشت نامعتبر است؛ متن الزامی است."))
		return
	}
	entityType, ok := parseNoteEntityType(in.EntityType)
	if !ok {
		httpx.HandleError(c, apperr.Validation("نوع پیوست یادداشت نامعتبر است."))
		return
	}

	note, err := h.services.Notes.Create(c.Request.Context(), services.CreateNoteInput{
		EntityType: entityType,
		EntityID:   in.EntityID,
		Body:       in.Body,
		Tags:       in.Tags,
		Pinned:     in.Pinned,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, note)
}

type updateNoteRequest struct {
	Body   *string  `json:"body"`
	Tags   []string `json:"tags"`
	Pinned *bool    `json:"pinned"`
}

func (h *FinancialHandlers) UpdateNote(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	var in updateNoteRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی بروزرسانی نامعتبر است."))
		return
	}
	note, err := h.services.Notes.Update(c.Request.Context(), id, services.UpdateNoteInput{
		Body:   in.Body,
		Tags:   in.Tags,
		Pinned: in.Pinned,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, note)
}

func (h *FinancialHandlers) DeleteNote(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Notes.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Backups ─────────────────────────────────────────────────────

type backupListResponse struct {
	Items          []services.BackupFile `json:"items"`
	LastAutoAt     *time.Time            `json:"last_auto_at"`
	IntervalHours  int                   `json:"interval_hours"`
}

func (h *FinancialHandlers) ListBackups(c *gin.Context) {
	items, err := h.backups.List()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	resp := backupListResponse{Items: items, IntervalHours: h.backupInterval()}
	if last, err := h.backups.LastAuto(c.Request.Context()); err == nil && !last.IsZero() {
		resp.LastAutoAt = &last
	}
	httpx.OK(c, http.StatusOK, resp)
}

func (h *FinancialHandlers) CreateBackup(c *gin.Context) {
	file, err := h.backups.Create(c.Request.Context(), false)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, file)
}

func (h *FinancialHandlers) DownloadBackup(c *gin.Context) {
	path, err := h.backups.Path(c.Param("name"))
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+c.Param("name")+"\"")
	c.File(path)
}

func (h *FinancialHandlers) DeleteBackup(c *gin.Context) {
	if err := h.backups.Delete(c.Param("name")); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FinancialHandlers) backupInterval() int { return backupIntervalHours }
