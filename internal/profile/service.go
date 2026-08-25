package profile

import (
	"context"
	"errors"
	"log"
	"my_feed/internal/account"
	"my_feed/internal/cache"
	"my_feed/internal/dberr"
	"time"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountRequired = errors.New("account required")
)

type ProfileService struct {
	accountRepo AccountReader
	videoRepo   VideoStatsReader
	followRepo  FollowStatsReader
	cache       ProfileCache
}

type AccountReader interface {
	FindAccountByID(id uint) (*account.Account, error)
}

type VideoStatsReader interface {
	GetVideosCount(accountID uint) (int64, error)
	GetAuthorLikesCount(accountID uint) (int64, error)
}

type FollowStatsReader interface {
	GetFollowersCount(accountID uint) (int64, error)
	GetFollowingsCount(accountID uint) (int64, error)
}

const profileCacheTTL = time.Minute

func NewProfileService(
	aR AccountReader,
	vR VideoStatsReader,
	fR FollowStatsReader,
	profileCache ProfileCache,
) *ProfileService {
	return &ProfileService{
		accountRepo: aR,
		videoRepo:   vR,
		followRepo:  fR,
		cache:       profileCache,
	}
}

// getProfile 返回格式
type ProfileAccount struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

type ProfileStats struct {
	VideosCount     int64 `json:"videos_count"`
	LikesCount      int64 `json:"likes_count"`
	FollowersCount  int64 `json:"followers_count"`
	FollowingsCount int64 `json:"followings_count"`
}

type Profile struct {
	Account *ProfileAccount `json:"account"`
	Stats   *ProfileStats   `json:"stats"`
}

func (service *ProfileService) GetProfile(ctx context.Context, accountID uint) (*Profile, error) {
	if accountID == 0 {
		return nil, ErrAccountRequired
	}

	if service.cache != nil {
		getCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		cached, hit, err := service.cache.GetProfile(getCtx, accountID)
		cancel()
		if err == nil {
			if hit {
				return &cached, nil
			}
		} else if !errors.Is(err, cache.ErrDisabled) {
			log.Printf("level=WARN component=profile_cache operation=get account_id=%d err=%q", accountID, err)
		}
	}

	// 查询用户信息
	account, err := service.accountRepo.FindAccountByID(accountID)
	if err != nil {
		if errors.Is(err, dberr.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	// 获取视频数
	videosCnt, err := service.videoRepo.GetVideosCount(accountID)
	if err != nil {
		return nil, err
	}
	// 获取获赞数
	likesCnt, err := service.videoRepo.GetAuthorLikesCount(accountID)
	if err != nil {
		return nil, err
	}
	// 获取粉丝数
	followersCnt, err := service.followRepo.GetFollowersCount(accountID)
	if err != nil {
		return nil, err
	}
	// 获取关注数
	followingsCnt, err := service.followRepo.GetFollowingsCount(accountID)
	if err != nil {
		return nil, err
	}

	profileAccount := &ProfileAccount{
		ID:        account.ID,
		Username:  account.Username,
		AvatarURL: account.AvatarURL,
		Bio:       account.Bio,
	}

	profileStats := &ProfileStats{
		VideosCount:     videosCnt,
		LikesCount:      likesCnt,
		FollowersCount:  followersCnt,
		FollowingsCount: followingsCnt,
	}

	profile := &Profile{
		Account: profileAccount,
		Stats:   profileStats,
	}

	if service.cache != nil {
		setCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		err = service.cache.SetProfile(setCtx, accountID, *profile, profileCacheTTL)
		cancel()
		if err != nil && !errors.Is(err, cache.ErrDisabled) {
			log.Printf("level=WARN component=profile_cache operation=set account_id=%d err=%q", accountID, err)
		}
	}

	return profile, nil
}
