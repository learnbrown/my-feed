package comment

import (
	"errors"
	"my_feed/internal/db"
	"my_feed/internal/video"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrAuthorRequired  = errors.New("author required")
	ErrVideoRequired   = errors.New("video required")
	ErrContentRequired = errors.New("comment content required")
	ErrContentTooLarge = errors.New("content too large")
	ErrCommentNotFound = errors.New("comment not found")
	ErrVideoNotFound   = errors.New("video not found")
	ErrCommentRequired = errors.New("comment required")
	ErrInvalidCursor   = errors.New("invalid cursor")
)

type CommentService struct {
	repo *CommentRepo
}

func NewCommentService(repo *CommentRepo) *CommentService {
	return &CommentService{repo: repo}
}

type CommentDTO struct {
	ID        uint   `json:"id"`
	VideoID   uint   `json:"video_id"`
	AccountID uint   `json:"account_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func (service *CommentService) CreateComment(accountID, videoID uint, content string) (*CommentDTO, uint, error) {
	if accountID == 0 {
		return nil, 0, ErrAuthorRequired
	}
	if videoID == 0 {
		return nil, 0, ErrVideoRequired
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, 0, ErrContentRequired
	}

	if utf8.RuneCountInString(content) > 500 {
		return nil, 0, ErrContentTooLarge
	}

	comment := &Comment{}
	var commentsCount uint
	err := service.repo.Transaction(func(commentRepo *CommentRepo, videoRepo *video.VideoRepo) error {
		_, err := videoRepo.FindVideoByID(videoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		cmt, err := commentRepo.CreateComment(accountID, videoID, content)
		if err != nil {
			return err
		}

		comment = cmt

		err = videoRepo.IncreaseCommentsCount(videoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		cnt, err := videoRepo.GetCommentsCount(videoID)
		if err != nil {
			return err
		}

		commentsCount = cnt

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	dto := &CommentDTO{
		ID:        comment.ID,
		VideoID:   comment.VideoID,
		AccountID: comment.AccountID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.UnixMilli(),
	}

	return dto, commentsCount, nil
}

func (service *CommentService) DeleteComment(accountID, commentID uint) (uint, error) {
	if accountID == 0 {
		return 0, ErrAuthorRequired
	}
	if commentID == 0 {
		return 0, ErrCommentRequired
	}

	var commentsCount uint

	err := service.repo.Transaction(func(commentRepo *CommentRepo, videoRepo *video.VideoRepo) error {
		comment, err := commentRepo.FindCommentByID(commentID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrCommentNotFound
			}
			return err
		}

		err = commentRepo.DeleteComment(accountID, commentID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrCommentNotFound
			}
			return err
		}

		err = videoRepo.DecreaseCommentsCount(comment.VideoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		cnt, err := videoRepo.GetCommentsCount(comment.VideoID)
		if err != nil {
			return err
		}

		commentsCount = cnt

		return nil
	})

	if err != nil {
		return 0, err
	}

	return commentsCount, nil
}

// listComment返回类型
type ListCommentResponse struct {
	Comments []CommentDTO `json:"comments"`
	HasMore  bool         `json:"has_more"`
	NextTime int64        `json:"next_time"`
	NextID   uint         `json:"next_id"`
}

func (service *CommentService) ListComment(videoID uint, limit int, latestTime time.Time, latestID uint) (*ListCommentResponse, error) {
	// TODO: 验证videoID有效性，先不管，没有该视频就返回空列表

	if videoID == 0 {
		return nil, ErrVideoRequired
	}
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	if latestTime.IsZero() != (latestID == 0) {
		return nil, ErrInvalidCursor
	}

	comments, err := service.repo.ListComment(videoID, limit+1, latestTime, latestID)
	if err != nil {
		return nil, err
	}

	hasMore := len(*comments) > limit
	if hasMore {
		*comments = (*comments)[:limit]
	}

	var nextTime int64
	var nextID uint
	if len(*comments) > 0 {
		nextTime = (*comments)[len(*comments)-1].CreatedAt.UnixMilli()
		nextID = (*comments)[len(*comments)-1].ID
	}

	dtos := make([]CommentDTO, len(*comments))
	for i, c := range *comments {
		dtos[i] = CommentDTO{
			ID:        c.ID,
			VideoID:   c.VideoID,
			AccountID: c.AccountID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt.UnixMilli(),
		}
	}

	return &ListCommentResponse{
		Comments: dtos,
		HasMore:  hasMore,
		NextTime: nextTime,
		NextID:   nextID,
	}, nil
}
