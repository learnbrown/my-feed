package video

import (
	"time"

	"gorm.io/gorm"
)

type VideoRepo struct {
	db *gorm.DB
}

func NewVideoRepo(db *gorm.DB) *VideoRepo {
	return &VideoRepo{db: db}
}

func (repo *VideoRepo) CreateVideo(video *Video) error {
	err := repo.db.Create(video).Error
	return err
}

func (repo *VideoRepo) FindVideoByID(id uint) (*Video, error) {
	video := &Video{}
	err := repo.db.Where("status = ?", 1).First(video, id).Error
	return video, err
}

// 分页查询，latest_time为游标，查询早于它创建的视频
func (repo *VideoRepo) ListByAuthorID(authorID uint, limit int, latestTime time.Time) (*[]Video, error) {
	videos := &[]Video{}

	query := repo.db.Model(&Video{}).Where("author_id = ? AND status = ?", authorID, 1)

	if !latestTime.IsZero() {
		query = query.Where("created_at < ?", latestTime)
	}

	err := query.Order("created_at desc").Limit(limit).Find(videos).Error
	return videos, err
}

// TODO: ListLatest 首页最新视频流
func (repo *VideoRepo) ListLatest() {

}
