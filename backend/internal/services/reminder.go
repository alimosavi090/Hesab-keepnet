package services

import (
	"context"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"gorm.io/gorm"
)

type ReminderService struct {
	db *gorm.DB
}

type CreateReminderInput struct {
	Title          string               `json:"title"`
	Description    *string              `json:"description"`
	DueDate        time.Time            `json:"due_date"`
	RepeatInterval enums.RepeatInterval `json:"repeat_interval"`
}

func (s *ReminderService) Create(ctx context.Context, in CreateReminderInput) (*models.Reminder, error) {
	if in.Title == "" {
		return nil, apperr.Validation("عنوان یادآور الزامی است.")
	}
	if !in.RepeatInterval.Valid() {
		in.RepeatInterval = enums.RepeatNone
	}

	reminder := models.Reminder{
		Title:          in.Title,
		Description:    in.Description,
		DueDate:        normalizeUTC(in.DueDate),
		RepeatInterval: in.RepeatInterval,
		IsDone:         false,
	}
	if err := s.db.WithContext(ctx).Create(&reminder).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return &reminder, nil
}

func (s *ReminderService) Get(ctx context.Context, id int64) (*models.Reminder, error) {
	var reminder models.Reminder
	if err := s.db.WithContext(ctx).First(&reminder, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &reminder, nil
}

func (s *ReminderService) List(ctx context.Context) ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := s.db.WithContext(ctx).Order("due_date").Find(&reminders).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return reminders, nil
}

func (s *ReminderService) Upcoming(ctx context.Context, within time.Duration) ([]models.Reminder, error) {
	now := time.Now().UTC()
	until := now.Add(within)
	var reminders []models.Reminder
	if err := s.db.WithContext(ctx).
		Where("is_done = ? AND due_date BETWEEN ? AND ?", false, now.Add(-72*time.Hour), until).
		Order("due_date").
		Find(&reminders).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return reminders, nil
}

func (s *ReminderService) MarkDone(ctx context.Context, id int64, done bool) error {
	result := s.db.WithContext(ctx).Model(&models.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_done":      done,
			"completed_at": completedAt(done),
		})
	if result.Error != nil {
		return apperr.Database(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("یادآور یافت نشد.")
	}
	return nil
}

func (s *ReminderService) Delete(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Delete(&models.Reminder{}, id)
	if result.Error != nil {
		return apperr.Database(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("یادآور یافت نشد.")
	}
	return nil
}

func completedAt(done bool) any {
	if done {
		return gorm.Expr("datetime('now')")
	}
	return nil
}
