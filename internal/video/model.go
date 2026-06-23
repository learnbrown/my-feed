package video

import (
	"time"

	"gorm.io/gorm"
)

type Video struct {
	gorm.Model

	AuthorID uint   `gorm:"index:idx_videos_author_created_at,priority:1;not null" json:"author_id"`
	Title    string `gorm:"size:128;not null" json:"title"`

	// 视频描述 可选；后续可以从这里提取 `#话题`
	Description string `gorm:"type:text" json:"description"`

	PlayURL string `gorm:"size:255;not null" json:"play_url"`

	// [x] 封面可选，如果未上传封面，使用默认封面
	CoverURL string `gorm:"size:255;default:'/static/uploads/covers/default.png'" json:"cover_url"`

	// 点赞数冗余字段 V0.2 固定为 0，V0.3 点赞时维护
	LikesCount uint `gorm:"default:0" json:"likes_count"`
	// 评论数冗余字段 V0.2 固定为 0，V0.3 评论时维护
	CommentsCount uint `gorm:"default:0" json:"comments_count"`
	// 热度分 V0.2 固定为 0，V1.0 热榜会用
	Popularity uint `gorm:"default:0" json:"popularity"`
	// 视频状态 `1` 表示正常；以后可扩展为审核中、删除、封禁
	Status int `gorm:"default:1" json:"status"`

	// 显式覆盖gorm.Model 创建联合索引
	CreatedAt time.Time `gorm:"index:idx_videos_author_created_at,priority:2;index:idx_videos_created_at"`
}

type Tag struct {
	gorm.Model
	Name string `gorm:"uniqueIndex:uk_tags_name;not null" json:"name"`
}

// [x] 怎么实现unique(video_id, tag_id) -> uniqueIndex:uk_video_tag
type VideoTag struct {
	gorm.Model
	VideoID uint `gorm:"not null;uniqueIndex:uk_video_tag" json:"video_id"`
	TagID   uint `gorm:"not null;uniqueIndex:uk_video_tag" json:"tag_id"`
}
