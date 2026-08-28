package services

import (
	"context"
	"sort"
	"strings"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/repository"
	"gorm.io/gorm"
)

type NotesService struct {
	db    *gorm.DB
	audit repository.AuditRepository
}

func (s *NotesService) normalizeTags(tags []string) string {
	seen := map[string]bool{}
	clean := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" || seen[t] || len(t) > 32 {
			continue
		}
		seen[t] = true
		clean = append(clean, strings.ReplaceAll(t, ",", ""))
	}
	sort.Strings(clean)
	return strings.Join(clean, ",")
}

type CreateNoteInput struct {
	EntityType models.NoteEntityType `json:"entity_type"`
	EntityID   *int64                `json:"entity_id"`
	Body       string                `json:"body"`
	Tags       []string              `json:"tags"`
	Pinned     bool                  `json:"pinned"`
}

var validNoteEntities = map[models.NoteEntityType]bool{
	models.NoteEntityRepresentative: true,
	models.NoteEntitySale:           true,
	models.NoteEntityBankAccount:    true,
	models.NoteEntityJournal:        true,
}

func (s *NotesService) Create(ctx context.Context, in CreateNoteInput) (*models.Note, error) {
	in.Body = strings.TrimSpace(in.Body)
	if in.Body == "" {
		return nil, apperr.Validation("متن یادداشت خالی است.")
	}
	if len(in.Body) > 10_000 {
		return nil, apperr.Validation("یادداشت بیش از حد بلند است.")
	}
	if !validNoteEntities[in.EntityType] {
		return nil, apperr.Validation("نوع پیوست یادداشت نامعتبر است.")
	}
	if in.EntityType == models.NoteEntityJournal {
		in.EntityID = nil // journal notes are not attached to any record
	} else if in.EntityID == nil || *in.EntityID <= 0 {
		return nil, apperr.Validation("شناسه رکورد برای این نوع یادداشت الزامی است.")
	}

	note := models.Note{
		EntityType: in.EntityType,
		EntityID:   in.EntityID,
		Body:       in.Body,
		Tags:       s.normalizeTags(in.Tags),
		Pinned:     in.Pinned,
	}
	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		if err := tx.Create(&note).Error; err != nil {
			return apperr.Database(err)
		}
		return writeAudit(s.audit, tx, ActionCreate, "note", note.ID, map[string]any{
			"entity_type": string(note.EntityType),
			"entity_id":   note.EntityID,
		})
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

type UpdateNoteInput struct {
	Body   *string  `json:"body"`
	Tags   []string `json:"tags"`
	Pinned *bool    `json:"pinned"`
}

func (s *NotesService) Update(ctx context.Context, id int64, in UpdateNoteInput) (*models.Note, error) {
	var note models.Note
	if err := s.db.WithContext(ctx).First(&note, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	if in.Body != nil {
		body := strings.TrimSpace(*in.Body)
		if body == "" {
			return nil, apperr.Validation("متن یادداشت خالی است.")
		}
		note.Body = body
	}
	if in.Tags != nil {
		note.Tags = s.normalizeTags(in.Tags)
	}
	if in.Pinned != nil {
		note.Pinned = *in.Pinned
	}
	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		if err := tx.Save(&note).Error; err != nil {
			return apperr.Database(err)
		}
		return writeAudit(s.audit, tx, ActionUpdate, "note", note.ID, map[string]any{
			"pinned": note.Pinned,
		})
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

type NoteFilter struct {
	EntityType models.NoteEntityType
	EntityID   *int64
	JournalOnly bool
	Pinned     *bool
	Tag        string
	Query      string
	PageQuery  PageQuery
}

func (f NoteFilter) Normalized() NoteFilter { return f }

// List returns notes newest-first; pinned journal entries float on top.
func (s *NotesService) List(ctx context.Context, filter NoteFilter) ([]models.Note, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Note{}).
		Where("deleted_at IS NULL")

	if filter.JournalOnly {
		query = query.Where("entity_type = ?", models.NoteEntityJournal)
	} else if filter.EntityType != "" {
		if !validNoteEntities[filter.EntityType] {
			return nil, 0, apperr.Validation("نوع پیوست یادداشت نامعتبر است.")
		}
		query = query.Where("entity_type = ?", filter.EntityType)
		if filter.EntityID != nil && *filter.EntityID > 0 {
			query = query.Where("entity_id = ?", *filter.EntityID)
		}
	}

	if filter.Pinned != nil {
		query = query.Where("pinned = ?", boolToInt(*filter.Pinned))
	}
	if tag := strings.TrimSpace(filter.Tag); tag != "" {
		tag = strings.ToLower(tag)
		query = query.Where("(tags = ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?)",
			tag, tag+",%", "%,"+tag, "%,"+tag+",%")
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + q + "%"
		query = query.Where("body LIKE ? OR tags LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	offset, limit := filter.PageQuery.Normalized()
	var notes []models.Note
	if err := query.Order("pinned DESC, created_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&notes).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}
	return notes, total, nil
}

func (s *NotesService) Delete(ctx context.Context, id int64) error {
	return database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var note models.Note
		if err := tx.First(&note, id).Error; err != nil {
			return apperr.Normalize(err)
		}
		if err := tx.Delete(&note).Error; err != nil {
			return apperr.Database(err)
		}
		return writeAudit(s.audit, tx, ActionDelete, "note", id, map[string]any{
			"entity_type": string(note.EntityType),
		})
	})
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
