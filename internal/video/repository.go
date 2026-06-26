package video

import (
	"errors"
	"my_feed/internal/db"
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

// TODO: ListLatest 首页最新视频流 升级为created_at, id 复合游标
func (repo *VideoRepo) ListLatest(limit int, latestTime time.Time) (*[]Video, error) {
	videos := &[]Video{}

	query := repo.db.Model(&Video{}).Where("status = ?", 1)

	if !latestTime.IsZero() {
		query = query.Where("created_at < ?", latestTime)
	}

	err := query.Order("created_at desc").
		Limit(limit).
		Find(videos).Error

	return videos, err
}

func (repo *VideoRepo) FindOrCreateTag(name string) (*Tag, error) {
	tag := &Tag{}

	// 先查询，若无再创建
	err := repo.db.Where("name = ?", name).First(tag).Error
	if err == nil {
		return tag, nil
	}
	if !errors.Is(err, db.ErrRecordNotFound) {
		return nil, err
	}

	tag = &Tag{Name: name}

	err = repo.db.Create(tag).Error

	if err == nil {
		return tag, nil
	}

	// 处理唯一索引冲突，防止并发创建同一个tag
	if db.IsDuplicateKeyError(err) {
		err = repo.db.Where("name = ?", name).First(tag).Error
		return tag, err
	}

	return nil, err
}

func (repo *VideoRepo) CreateVideoTag(videoID, tagID uint) error {
	videoTag := &VideoTag{
		VideoID: videoID,
		TagID:   tagID,
	}
	err := repo.db.Create(videoTag).Error
	return err
}

// 用于将video、tag、video-tag的创建放入gorm事务，统一返回错误并回滚
func (repo *VideoRepo) Transaction(fn func(txRepo *VideoRepo) error) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		txRepo := NewVideoRepo(tx)
		return fn(txRepo)
	})
}

func (repo *VideoRepo) ListByTag(tagName string, limit int, latestTime time.Time) (*[]Video, error) {
	videos := &[]Video{}

	tagQuery := repo.db.Model(&Tag{}).Select("id").Where("name = ?", tagName)

	videoTagQuery := repo.db.Model(&VideoTag{}).Select("video_id").Where("tag_id IN (?)", tagQuery)

	videoQuery := repo.db.Model(&Video{}).
		Where("id IN (?)", videoTagQuery).
		Where("status = ?", 1)

	if !latestTime.IsZero() {
		videoQuery = videoQuery.Where("created_at < ?", latestTime)
	}

	err := videoQuery.Order("created_at desc").
		Limit(limit).
		Find(videos).Error

	return videos, err
}

// [x] 点赞数应该由谁获取？就你
func (repo *VideoRepo) IncreaseLikesCount(id uint) error {
	res := repo.db.Model(&Video{}).
		Where("id = ? AND status = 1", id).
		UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1))

	if res.Error != nil {
		return res.Error
	}
	// 还应检查RowsAffected，如果视频不存在、状态不对、或者 likes_count > 0 条件没命中，Error 仍可能是 nil
	if res.RowsAffected == 0 {
		return db.ErrRecordNotFound
	}
	return nil
}

func (repo *VideoRepo) DecreaseLikesCount(id uint) error {
	res := repo.db.Model(&Video{}).
		Where("id = ? AND likes_count > 0 AND status = 1", id).
		UpdateColumn("likes_count", gorm.Expr("likes_count - ?", 1))

	if res.Error != nil {
		return res.Error
	}
	// TODO: RowsAffected = 0 不一定是视频不存在，有三种情况：
	// 视频不存在
	// 视频 status != 1
	// likes_count 已经是 0
	// 通常情况是likes_count 已经是 0，后面改成更明确的错误
	if res.RowsAffected == 0 {
		return db.ErrRecordNotFound
	}

	return nil
}

// 查询视频获赞数
func (repo *VideoRepo) GetLikesCount(id uint) (uint, error) {
	video := &Video{}
	err := repo.db.First(video, id).Error

	return video.LikesCount, err
}

// TODO: 维护评论数字段
func (repo *VideoRepo) IncreaseCommentsCount(id uint) error {
	res := repo.db.Model(&Video{}).
		Where("id = ? AND status = 1", id).
		UpdateColumn("comments_count", gorm.Expr("comments_count + ?", 1))

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return db.ErrRecordNotFound
	}
	return nil
}

func (repo *VideoRepo) DecreaseCommentsCount(id uint) error {
	res := repo.db.Model(&Video{}).
		Where("id = ? AND comments_count > 0 AND status = 1", id).
		UpdateColumn("comments_count", gorm.Expr("comments_count - ?", 1))

	if res.Error != nil {
		return res.Error
	}
	// TODO: RowsAffected = 0 不一定是视频不存在，有三种情况：
	// 视频不存在
	// 视频 status != 1
	// comments_count 已经是 0
	// 通常情况是comments_count 已经是 0，后面改成更明确的错误
	if res.RowsAffected == 0 {
		return db.ErrRecordNotFound
	}

	return nil
}

func (repo *VideoRepo) GetCommentsCount(id uint) (uint, error) {
	video := &Video{}
	err := repo.db.First(video, id).Error

	return video.CommentsCount, err
}

// 查询某用户上传视频数
func (repo *VideoRepo) GetVideosCount(accountID uint) (int64, error) {
	var cnt int64
	err := repo.db.Model(&Video{}).
		Where("author_id = ? AND status = ?", accountID, 1).
		Count(&cnt).Error
	return cnt, err
}

// 查询某用户视频获赞数
func (repo *VideoRepo) GetAuthorLikesCount(accountID uint) (int64, error) {
	var cnt int64
	err := repo.db.Model(&Video{}).
		Where("author_id = ? AND status = ?", accountID, 1).
		Select("COALESCE(SUM(likes_count), 0)").
		Scan(&cnt).Error
	return cnt, err
}
