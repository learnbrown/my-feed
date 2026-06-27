package like

import (
	"errors"
	"my_feed/internal/db"
	"my_feed/internal/video"
	"time"
)

var (
	ErrAuthorRequired = errors.New("author required") // 未登录，无法获取当前用户的id
	ErrVideoRequired  = errors.New("video required")
	ErrVideoNotFound  = errors.New("video not found")
	ErrLikeNotFound   = errors.New("like not found")
)

type LikeService struct {
	repo *LikeRepo
}

func NewLikeService(repo *LikeRepo) *LikeService {
	return &LikeService{repo: repo}
}

func (service *LikeService) Like(accountID, videoID uint) (uint, error) {
	if accountID == 0 {
		return 0, ErrAuthorRequired
	}
	if videoID == 0 {
		return 0, ErrVideoRequired
	}

	var likes uint

	err := service.repo.Transaction(func(likeRepo *LikeRepo, videoRepo *video.VideoRepo) error {
		// 查询视频是否存在
		_, err := videoRepo.FindVideoByID(videoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		// 创建点赞记录
		err = likeRepo.CreateLike(accountID, videoID)
		if err != nil {
			// 唯一索引冲突，查询likes_count并返回成功
			if db.IsDuplicateKeyError(err) {
				cnt, err := videoRepo.GetLikesCount(videoID)
				if err != nil {
					return err
				}
				likes = cnt
				return nil
			}
			return err
		}

		// 更新video字段
		err = videoRepo.IncreaseLikesCount(videoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		// 获取点赞数
		cnt, err := videoRepo.GetLikesCount(videoID)
		if err != nil {
			return err
		}
		likes = cnt

		return nil
	})

	return likes, err
}

func (service *LikeService) Unlike(accountID, videoID uint) (uint, error) {
	if accountID == 0 {
		return 0, ErrAuthorRequired
	}
	if videoID == 0 {
		return 0, ErrVideoRequired
	}

	var likes uint

	err := service.repo.Transaction(func(likeRepo *LikeRepo, videoRepo *video.VideoRepo) error {
		// 查看视频是否存在
		_, err := videoRepo.FindVideoByID(videoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		// 删除点赞记录
		deleted, err := likeRepo.DeleteLike(accountID, videoID)
		if err != nil {
			return err
		}
		if !deleted {
			cnt, err := videoRepo.GetLikesCount(videoID)
			if err != nil {
				return err
			}
			likes = cnt
			return nil
		}

		// 更新video字段
		err = videoRepo.DecreaseLikesCount(videoID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}

		// 获取点赞数
		cnt, err := videoRepo.GetLikesCount(videoID)
		if err != nil {
			return err
		}

		likes = cnt

		return nil
	})

	return likes, err
}

// ListLikedVideos repo层返回格式
// LikedVideos以likes.created_at排序
// LikeAt字段用于传递likes.created_at，以得到next_time
type LikedList struct {
	video.Video
	LikedAt time.Time `json:"liked_at"`
}

type LikedListDTO struct {
	ID            uint   `json:"id"`
	AuthorID      uint   `json:"author_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	PlayURL       string `json:"play_url"`
	CoverURL      string `json:"cover_url"`
	LikesCount    uint   `json:"likes_count"`
	CommentsCount uint   `json:"comments_count"`
	CreatedAt     int64  `json:"created_at"`
	LikedAt       int64  `json:"liked_at"`
}

func ToDTOs(likedList []LikedList) []LikedListDTO {
	dtos := make([]LikedListDTO, len(likedList))
	for i, v := range likedList {
		dtos[i] = LikedListDTO{
			ID:            v.ID,
			AuthorID:      v.AuthorID,
			Title:         v.Title,
			Description:   v.Description,
			PlayURL:       v.PlayURL,
			CoverURL:      v.CoverURL,
			LikesCount:    v.LikesCount,
			CommentsCount: v.CommentsCount,
			CreatedAt:     v.CreatedAt.UnixMilli(),
			LikedAt:       v.LikedAt.UnixMilli(),
		}
	}
	return dtos
}

type ListLikedResponse struct {
	HasMore  bool            `json:"has_more"`
	NextTime int64           `json:"next_time"`
	Likes    []LikedListDTO `json:"likes"`
}

func (service *LikeService) ListLikedVideos(accountID uint, limit int, latestTime time.Time) (*ListLikedResponse, error) {
	if accountID == 0 {
		return nil, ErrAuthorRequired
	}
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	likedList, err := service.repo.ListLikedVideos(accountID, limit+1, latestTime)
	if err != nil {
		return nil, err
	}

	hasMore := len(*likedList) > limit
	if hasMore {
		*likedList = (*likedList)[:limit]
	}

	var nextTime int64
	if len(*likedList) > 0 {
		nextTime = (*likedList)[len(*likedList)-1].LikedAt.UnixMilli()
	}

	dtos := ToDTOs(*likedList)

	res := &ListLikedResponse{
		Likes:    dtos,
		NextTime: nextTime,
		HasMore:  hasMore,
	}

	return res, err
}

func (service *LikeService) IsLiked(accountID, videoID uint) (bool, error) {
	if accountID == 0 {
		return false, ErrAuthorRequired
	}
	if videoID == 0 {
		return false, ErrVideoRequired
	}

	return service.repo.ExistsLike(accountID, videoID)
}
