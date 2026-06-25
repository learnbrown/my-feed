package follow

import "time"

type Follow struct {
	ID uint `gorm:"primaryKey"`

	// 粉丝ID
	FollowerID uint `gorm:"uniqueIndex:uk_follows_follower_vlogger,priority:1;not null" json:"follower_id"`

	// 博主ID
	VloggerID uint `gorm:"uniqueIndex:uk_follows_follower_vlogger,priority:2;index:idx_follows_vlogger_id;not null" json:"vlogger_id"`

	CreatedAt time.Time
}
