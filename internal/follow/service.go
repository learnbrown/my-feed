package follow

import (
	"errors"
	"my_feed/internal/account"
	"my_feed/internal/db"
	"time"
)

var (
	ErrFollowerRequired = errors.New("follower_id required")
	ErrVloggerRequired  = errors.New("vlogger_id required")
	ErrSelfFollowing    = errors.New("can't follow yourself")
	ErrVloggerNotFound  = errors.New("vlogger not found")
	ErrAccountNotFound  = errors.New("account not found")
)

// listFollower与listFollowing repo层向service返回格式
// 粉丝/博主信息与关注时间
type FollowList struct {
	account.Account
	FollowedAt time.Time `json:"followed_at"`
}

type FollowService struct {
	followRepo  *FollowRepo
	accountRepo *account.AccountRepo
}

func NewFollowService(followRepo *FollowRepo, accountRepo *account.AccountRepo) *FollowService {
	return &FollowService{followRepo: followRepo, accountRepo: accountRepo}
}

func (service *FollowService) IsFollowing(followerID, vloggerID uint) (bool, error) {
	if followerID == 0 {
		return false, ErrFollowerRequired
	}
	if vloggerID == 0 {
		return false, ErrVloggerRequired
	}

	// 对于自己直接返回未关注
	if followerID == vloggerID {
		return false, nil
	}

	return service.followRepo.ExistsFollow(followerID, vloggerID)
}

func (service *FollowService) Follow(followerID, vloggerID uint) error {
	if followerID == 0 {
		return ErrFollowerRequired
	}
	if vloggerID == 0 {
		return ErrVloggerRequired
	}

	if followerID == vloggerID {
		return ErrSelfFollowing
	}

	// 查看博主是否存在
	_, err := service.accountRepo.FindAccountByID(vloggerID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return ErrVloggerNotFound
		}
		return err
	}

	err = service.followRepo.CreateFollow(followerID, vloggerID)
	if err == nil || db.IsDuplicateKeyError(err) {
		return nil
	}

	return err
}

func (service *FollowService) Unfollow(followerID, vloggerID uint) error {
	if followerID == 0 {
		return ErrFollowerRequired
	}
	if vloggerID == 0 {
		return ErrVloggerRequired
	}

	if followerID == vloggerID {
		return ErrSelfFollowing
	}

	_, err := service.followRepo.DeleteFollow(followerID, vloggerID)
	if err != nil {
		return err
	}

	return nil
}

// listFollower/listFollowing 返回类型
type ListFollowResponse struct {
	Follows  *[]FollowList `json:"follows"` // 博主/粉丝列表
	HasMore  bool          `json:"has_more"`
	NextTime int64         `json:"next_time"`
}

func (service *FollowService) ListFollower(vloggerID uint, limit int, latestTime time.Time) (*ListFollowResponse, error) {
	if vloggerID == 0 {
		return nil, ErrVloggerRequired
	}
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	// 检查用户是否存在
	_, err := service.accountRepo.FindAccountByID(vloggerID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	followerList, err := service.followRepo.ListFollower(vloggerID, limit+1, latestTime)
	if err != nil {
		return nil, err
	}

	hasMore := len(*followerList) > limit
	if hasMore {
		*followerList = (*followerList)[:limit]
	}

	var nextTime int64
	if len(*followerList) > 0 {
		nextTime = (*followerList)[len(*followerList)-1].FollowedAt.UnixMilli()
	}

	return &ListFollowResponse{
		Follows:  followerList,
		HasMore:  hasMore,
		NextTime: nextTime,
	}, nil
}

func (service *FollowService) ListFollowing(followerID uint, limit int, latestTime time.Time) (*ListFollowResponse, error) {
	if followerID == 0 {
		return nil, ErrFollowerRequired
	}
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	// 检查用户是否存在
	_, err := service.accountRepo.FindAccountByID(followerID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	followingList, err := service.followRepo.ListFollowing(followerID, limit+1, latestTime)
	if err != nil {
		return nil, err
	}

	hasMore := len(*followingList) > limit
	if hasMore {
		*followingList = (*followingList)[:limit]
	}

	var nextTime int64
	if len(*followingList) > 0 {
		nextTime = (*followingList)[len(*followingList)-1].FollowedAt.UnixMilli()
	}

	return &ListFollowResponse{
		Follows:  followingList,
		HasMore:  hasMore,
		NextTime: nextTime,
	}, nil
}
