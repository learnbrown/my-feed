package follow

import (
	"errors"
	"my_feed/internal/account"
	"my_feed/internal/dberr"
	"time"

	"gorm.io/gorm"
)

type FollowRepo struct {
	db *gorm.DB
}

func NewFollowRepo(db *gorm.DB) *FollowRepo {
	return &FollowRepo{db: db}
}

func (repo *FollowRepo) ExistsFollow(followerID, vloggerID uint) (bool, error) {
	err := repo.db.Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		First(&Follow{}).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, dberr.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (repo *FollowRepo) CreateFollow(followerID, vloggerID uint) error {
	err := repo.db.Create(&Follow{FollowerID: followerID, VloggerID: vloggerID}).Error
	return err
}

func (repo *FollowRepo) DeleteFollow(followerID, vloggerID uint) (bool, error) {
	res := repo.db.Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Delete(&Follow{})
	return res.RowsAffected > 0, res.Error
}

// 粉丝列表
func (repo *FollowRepo) ListFollower(vloggerID uint, limit int, latestTime time.Time, latestID uint) (*[]FollowList, error) {
	followerList := &[]FollowList{}
	query := repo.db.Model(&account.Account{}).
		Select("accounts.*, follows.created_at AS followed_at, follows.id AS follow_id").
		Joins("JOIN follows ON accounts.id = follows.follower_id").
		Where("follows.vlogger_id = ?", vloggerID)

	if !latestTime.IsZero() && latestID > 0 {
		query = query.Where(
			"(follows.created_at < ? OR (follows.created_at = ? AND follows.id < ?))",
			latestTime, latestTime, latestID,
		)
	}

	err := query.Order("follows.created_at DESC, follows.id DESC").
		Limit(limit).
		Scan(followerList).Error

	return followerList, err
}

// 关注列表
func (repo *FollowRepo) ListFollowing(followerID uint, limit int, latestTime time.Time, latestID uint) (*[]FollowList, error) {
	followingList := &[]FollowList{}
	query := repo.db.Model(&account.Account{}).
		Select("accounts.*, follows.created_at AS followed_at, follows.id AS follow_id").
		Joins("JOIN follows ON accounts.id = follows.vlogger_id").
		Where("follows.follower_id = ?", followerID)

	if !latestTime.IsZero() && latestID > 0 {
		query = query.Where(
			"(follows.created_at < ? OR (follows.created_at = ? AND follows.id < ?))",
			latestTime, latestTime, latestID,
		)
	}
	err := query.Order("follows.created_at DESC, follows.id DESC").
		Limit(limit).
		Scan(followingList).Error
	return followingList, err
}

// 查看某用户的粉丝数
func (repo *FollowRepo) GetFollowersCount(accountID uint) (int64, error) {
	var cnt int64
	err := repo.db.Model(&Follow{}).
		Where("vlogger_id = ?", accountID).
		Count(&cnt).Error
	return cnt, err
}

// 查询某用户的关注数
func (repo *FollowRepo) GetFollowingsCount(accountID uint) (int64, error) {
	var cnt int64
	err := repo.db.Model(&Follow{}).
		Where("follower_id = ?", accountID).
		Count(&cnt).Error
	return cnt, err
}
