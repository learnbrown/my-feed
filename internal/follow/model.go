package follow

import "time"

type Follow struct {
	ID uint `gorm:"primary_key"`

	FollowerID uint `gorm:"uniqueIndex:uk_follows_follower_vlogger,priority:1;not null" json:"follower_id"`
	VloggerID  uint `gorm:"uniqueIndex:uk_follows_follower_vlogger,priority:2;index:idx_follows_vlogger_id;not null" json:"vlogger_id"`

	CreatedAt time.Time
}
