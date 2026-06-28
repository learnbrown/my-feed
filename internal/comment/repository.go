package comment

import (
	"my_feed/internal/dberr"
	"my_feed/internal/video"
	"time"

	"gorm.io/gorm"
)

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (repo *CommentRepo) Transaction(fn func(commentRepo *CommentRepo, videoRepo *video.VideoRepo) error) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		commentRepo := NewCommentRepo(tx)
		videoRepo := video.NewVideoRepo(tx)
		return fn(commentRepo, videoRepo)
	})
}

func (repo *CommentRepo) CreateComment(accountID, videoID uint, content string) (*Comment, error) {
	comment := &Comment{
		AccountID: accountID,
		VideoID:   videoID,
		Content:   content,
	}
	err := repo.db.Create(comment).Error
	return comment, err
}

// TODO: “不存在”和“不是你的评论”都变成 ErrCommentNotFound
// 如果想语义更精确，应该区分 ErrCommentForbidden 返回 403
func (repo *CommentRepo) DeleteComment(accountID, commentID uint) error {
	res := repo.db.Model(&Comment{}).
		Where("id = ? AND account_id = ?", commentID, accountID).
		Delete(&Comment{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return dberr.ErrRecordNotFound
	}

	return nil
}

func (repo *CommentRepo) FindCommentByID(commentID uint) (*Comment, error) {
	comment := &Comment{}
	err := repo.db.First(comment, commentID).Error

	return comment, err
}

func (repo *CommentRepo) ListComment(videoID uint, limit int, latestTime time.Time, latestID uint) (*[]Comment, error) {
	comments := &[]Comment{}
	query := repo.db.Model(&Comment{}).
		Where("video_id = ?", videoID)

	if !latestTime.IsZero() && latestID > 0 {
		query = query.Where(
			"(created_at < ? OR (created_at = ? AND id < ?))",
			latestTime, latestTime, latestID,
		)
	}

	err := query.Order("created_at DESC, id DESC").
		Limit(limit).
		Find(comments).Error

	return comments, err
}
