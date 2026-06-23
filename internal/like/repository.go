package like

import (
	"errors"
	"my_feed/internal/db"
	"my_feed/internal/video"
	"time"

	"gorm.io/gorm"
)

type LikeRepo struct {
	db *gorm.DB
}

func NewLikeRepo(db *gorm.DB) *LikeRepo {
	return &LikeRepo{db: db}
}

// gorm事务
func (repo *LikeRepo) Transaction(fn func(likeRepo *LikeRepo, videoRepo *video.VideoRepo) error) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		likeRepo := NewLikeRepo(tx)
		videoRepo := video.NewVideoRepo(tx)
		return fn(likeRepo, videoRepo)
	})
}

func (repo *LikeRepo) ExistsLike(accountID, videoID uint) (bool, error) {
	like := &Like{}
	err := repo.db.Model(&Like{}).
		Where("account_id = ? AND video_id = ?", accountID, videoID).
		First(like).Error

	if err == nil {
		return true, nil
	}
	if errors.Is(err, db.ErrRecordNotFound) {
		return false, nil
	}

	return false, err
}

func (repo *LikeRepo) CreateLike(accountID, videoID uint) error {
	err := repo.db.Model(&Like{}).
		Create(&Like{AccountID: accountID, VideoID: videoID}).
		Error

	return err
}

// [x] 删除点赞记录怎么得到点赞数？不需要它获取
func (repo *LikeRepo) DeleteLike(accountID, videoID uint) (deleted bool, err error) {
	res := repo.db.Model(&Like{}).
		Where("account_id = ? AND video_id = ?", accountID, videoID).
		Delete(&Like{})

	return res.RowsAffected > 0, res.Error
}

// 返回点赞过的视频列表
// [x] 应该按 likes.created_at desc 排序
// [x] 返回值中不包含likes.created_at，无法获得next_time -> 定义新结构来保存点赞时间
func (repo *LikeRepo) ListLikedVideos(accountID uint, limit int, latestTime time.Time) (*[]LikedList, error) {
	likedList := &[]LikedList{}

	videoQuery := repo.db.Model(&video.Video{}).
		Select("videos.*, likes.created_at AS liked_at").
		Joins("JOIN likes ON videos.id = likes.video_id").
		Where("likes.account_id = ? AND videos.status = 1", accountID)

	if !latestTime.IsZero() {
		videoQuery = videoQuery.Where("likes.created_at < ?", latestTime)
	}

	err := videoQuery.Order("likes.created_at DESC").
		Limit(limit).
		Scan(likedList).Error

	return likedList, err
}
