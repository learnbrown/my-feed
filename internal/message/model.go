package message

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID uint `gorm:"primaryKey"`

	FromID    uint      `gorm:"not null;index:idx_messages_from_to_created_at,priority:1;index:idx_messages_to_from_created_at,priority:2" json:"from_id"`
	ToID      uint      `gorm:"not null;index:idx_messages_from_to_created_at,priority:2;index:idx_messages_to_from_created_at,priority:1" json:"to_id"`
	Content   string    `gorm:"size:1000;not null" json:"content"`
	CreatedAt time.Time `gorm:"index:idx_messages_from_to_created_at,priority:3;index:idx_messages_to_from_created_at,priority:3"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
