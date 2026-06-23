package like

import (
	"time"
)

type Like struct {
	// gorm.Model
	// 如果有DeletedAt字段，gorm会采用软删除
	// 导致点赞记录被删但唯一索引依然存在，无法再次点赞

	ID uint `gorm:"primarykey"`

	AccountID uint `gorm:"uniqueIndex:uk_likes_video_account,priority:2;index:idx_likes_account_created_at,priority:1;not null" json:"account_id"`
	VideoID   uint `gorm:"uniqueIndex:uk_likes_video_account,priority:1;not null" json:"video_id"`

	CreatedAt time.Time `gorm:"index:idx_likes_account_created_at,priority:2"`
}
