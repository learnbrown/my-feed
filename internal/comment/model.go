package comment

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model

	VideoID   uint      `gorm:"index:idx_comments_video_created_at,priority:1;not null" json:"video_id"`
	AccountID uint      `gorm:"index:idx_comments_account_id;not null" json:"account_id"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	CreatedAt time.Time `gorm:"index:idx_comments_video_created_at,priority:2"`
}
