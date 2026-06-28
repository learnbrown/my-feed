package profile

import (
	"errors"
	"my_feed/internal/account"
	"my_feed/internal/dberr"
	"my_feed/internal/follow"
	"my_feed/internal/video"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountRequired = errors.New("account required")
)

type ProfileService struct {
	accountRepo *account.AccountRepo
	videoRepo   *video.VideoRepo
	followRepo  *follow.FollowRepo
}

func NewProfileService(
	aR *account.AccountRepo,
	vR *video.VideoRepo,
	fR *follow.FollowRepo,
) *ProfileService {
	return &ProfileService{
		accountRepo: aR,
		videoRepo:   vR,
		followRepo:  fR,
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

func (service *ProfileService) GetProfile(accountID uint) (*Profile, error) {
	if accountID == 0 {
		return nil, ErrAccountRequired
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

	return profile, nil
}
