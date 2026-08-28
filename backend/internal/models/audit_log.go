package models

import "time"

type AuditLog struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ActorUserID *int64    `json:"actor_user_id" gorm:"column:actor_user_id"`
	Action      string    `json:"action" gorm:"column:action;not null"`
	EntityType  string    `json:"entity_type" gorm:"column:entity_type;not null"`
	EntityID    int64     `json:"entity_id" gorm:"column:entity_id;not null"`
	Metadata    *string   `json:"metadata" gorm:"column:metadata"`
	IPAddress   *string   `json:"ip_address" gorm:"column:ip_address"`
	UserAgent   *string   `json:"user_agent" gorm:"column:user_agent"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (AuditLog) TableName() string { return "audit_logs" }
