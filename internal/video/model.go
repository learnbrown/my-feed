package video

import "gorm.io/gorm"

type Video struct {
	gorm.Model
	AuthorID uint   `gorm:"index;not null" json:"author_id"`
	Title    string `gorm:"size:128;not null" json:"title"`

	// 视频描述 可选；后续可以从这里提取 `#话题`
	Description string `gorm:"type:text" json:"description"`

	PlayURL  string `gorm:"size:255;not null" json:"play_url"`
	CoverURL string `gorm:"size:255;not null" json:"cover_url"`

	// 点赞数冗余字段 V0.2 固定为 0，V0.3 点赞时维护
	LikesCount uint `gorm:"default:0" json:"likes_count"`
	// 评论数冗余字段 V0.2 固定为 0，V0.3 评论时维护
	CommentsCount uint `gorm:"default:0" json:"comments_count"`
	// 热度分 V0.2 固定为 0，V1.0 热榜会用
	Popularity uint `gorm:"default:0" json:"popularity"`
	// 视频状态 `1` 表示正常；以后可扩展为审核中、删除、封禁
	Status bool `gorm:"default:1" json:"status"`
}

type Tag struct {
	gorm.Model
	Name string `gorm:"index;unique;not null" json:"name"`
}

// TODO: 怎么实现unique(video_id, tag_id)
type VideoTag struct {
	gorm.Model
	VideoID uint `gorm:"not null" json:"video_id"`
	TagID   uint `gorm:"not null" json:"tag_id"`
}
