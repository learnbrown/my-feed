package follow

import "gorm.io/gorm"

type Follow struct {
	gorm.Model

	FollowerID uint `gorm:"uniqueIndex:uk_follows_follower_vlogger,priority:1;not null" json:"follower_id"`
	VloggerID  uint `gorm:"uniqueIndex:uk_follows_follower_vlogger,priority:2;index:idx_follows_vlogger_id;not null" json:"vlogger_id"`
}
