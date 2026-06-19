package video

import "gorm.io/gorm"

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
	err := repo.db.First(video, id).Error
	return video, err
}

func (repo *VideoRepo) ListByAuthorID(id uint) (*[]Video, error) {
	videos := &[]Video{}
	err := repo.db.Model(&Video{}).Find(videos, id).Error
	return videos, err
}

// TODO: ListLatest 首页最新视频流
func (repo *VideoRepo) ListLatest() {

}