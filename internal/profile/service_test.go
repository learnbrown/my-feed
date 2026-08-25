package profile

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"my_feed/internal/account"
	"my_feed/internal/dberr"
)

type accountReaderStub struct {
	account *account.Account
	err     error
	calls   int
}

func (stub *accountReaderStub) FindAccountByID(uint) (*account.Account, error) {
	stub.calls++
	return stub.account, stub.err
}

type videoStatsReaderStub struct {
	videosCount int64
	likesCount  int64
	videosErr   error
	likesErr    error
	videosCalls int
	likesCalls  int
}

func (stub *videoStatsReaderStub) GetVideosCount(uint) (int64, error) {
	stub.videosCalls++
	return stub.videosCount, stub.videosErr
}

func (stub *videoStatsReaderStub) GetAuthorLikesCount(uint) (int64, error) {
	stub.likesCalls++
	return stub.likesCount, stub.likesErr
}

type followStatsReaderStub struct {
	followersCount  int64
	followingsCount int64
	followersErr    error
	followingsErr   error
	followersCalls  int
	followingsCalls int
}

func (stub *followStatsReaderStub) GetFollowersCount(uint) (int64, error) {
	stub.followersCalls++
	return stub.followersCount, stub.followersErr
}

func (stub *followStatsReaderStub) GetFollowingsCount(uint) (int64, error) {
	stub.followingsCalls++
	return stub.followingsCount, stub.followingsErr
}

type profileCacheStub struct {
	profile      Profile
	hit          bool
	getErr       error
	setErr       error
	getCalls     int
	setCalls     int
	setAccountID uint
	setProfile   Profile
	setTTL       time.Duration
}

func (stub *profileCacheStub) GetProfile(context.Context, uint) (Profile, bool, error) {
	stub.getCalls++
	return stub.profile, stub.hit, stub.getErr
}

func (stub *profileCacheStub) SetProfile(_ context.Context, accountID uint, profile Profile, ttl time.Duration) error {
	stub.setCalls++
	stub.setAccountID = accountID
	stub.setProfile = profile
	stub.setTTL = ttl
	return stub.setErr
}

func (stub *profileCacheStub) DelProfile(context.Context, uint) error {
	return nil
}

func TestGetProfileRequiresAccount(t *testing.T) {
	service := ProfileService{}

	if _, err := service.GetProfile(t.Context(), 0); !errors.Is(err, ErrAccountRequired) {
		t.Fatalf("GetProfile() error = %v, want ErrAccountRequired", err)
	}
}

func TestGetProfileReturnsCacheHitWithoutRepositories(t *testing.T) {
	want := Profile{
		Account: &ProfileAccount{ID: 9, Username: "cached"},
		Stats:   &ProfileStats{VideosCount: 2, LikesCount: 3, FollowersCount: 4, FollowingsCount: 5},
	}
	profileCache := &profileCacheStub{profile: want, hit: true}
	service := NewProfileService(nil, nil, nil, profileCache)

	got, err := service.GetProfile(t.Context(), want.Account.ID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("GetProfile() = %#v, want %#v", got, want)
	}
	if profileCache.getCalls != 1 || profileCache.setCalls != 0 {
		t.Fatalf("cache calls get=%d set=%d, want get=1 set=0", profileCache.getCalls, profileCache.setCalls)
	}
}

func TestGetProfileCacheMissAggregatesAndFillsCache(t *testing.T) {
	accountRepo := &accountReaderStub{account: &account.Account{Username: "alice", AvatarURL: "/avatar.png", Bio: "bio"}}
	accountRepo.account.ID = 12
	videoRepo := &videoStatsReaderStub{videosCount: 2, likesCount: 8}
	followRepo := &followStatsReaderStub{followersCount: 3, followingsCount: 4}
	profileCache := &profileCacheStub{}
	service := NewProfileService(accountRepo, videoRepo, followRepo, profileCache)

	got, err := service.GetProfile(t.Context(), accountRepo.account.ID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	want := Profile{
		Account: &ProfileAccount{ID: 12, Username: "alice", AvatarURL: "/avatar.png", Bio: "bio"},
		Stats:   &ProfileStats{VideosCount: 2, LikesCount: 8, FollowersCount: 3, FollowingsCount: 4},
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("GetProfile() = %#v, want %#v", got, want)
	}
	if accountRepo.calls != 1 || videoRepo.videosCalls != 1 || videoRepo.likesCalls != 1 || followRepo.followersCalls != 1 || followRepo.followingsCalls != 1 {
		t.Fatalf("unexpected repository calls: account=%d videos=%d likes=%d followers=%d followings=%d", accountRepo.calls, videoRepo.videosCalls, videoRepo.likesCalls, followRepo.followersCalls, followRepo.followingsCalls)
	}
	if profileCache.setCalls != 1 || profileCache.setAccountID != 12 || profileCache.setTTL != time.Minute || !reflect.DeepEqual(profileCache.setProfile, want) {
		t.Fatalf("unexpected cache fill: calls=%d accountID=%d ttl=%s profile=%#v", profileCache.setCalls, profileCache.setAccountID, profileCache.setTTL, profileCache.setProfile)
	}
}

func TestGetProfileCacheFailuresDoNotBlockDatabaseResult(t *testing.T) {
	tests := []struct {
		name   string
		getErr error
		setErr error
	}{
		{name: "get failure", getErr: errors.New("redis get unavailable")},
		{name: "set failure", setErr: errors.New("redis set unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &accountReaderStub{account: &account.Account{Username: "alice"}}
			accountRepo.account.ID = 12
			profileCache := &profileCacheStub{getErr: tt.getErr, setErr: tt.setErr}
			service := NewProfileService(accountRepo, &videoStatsReaderStub{}, &followStatsReaderStub{}, profileCache)

			got, err := service.GetProfile(t.Context(), accountRepo.account.ID)
			if err != nil {
				t.Fatalf("GetProfile() error = %v", err)
			}
			if got.Account.ID != accountRepo.account.ID || accountRepo.calls != 1 || profileCache.setCalls != 1 {
				t.Fatalf("expected database fallback and cache fill attempt, got=%#v accountCalls=%d setCalls=%d", got, accountRepo.calls, profileCache.setCalls)
			}
		})
	}
}

func TestGetProfileDoesNotCacheMissingAccount(t *testing.T) {
	accountRepo := &accountReaderStub{err: dberr.ErrRecordNotFound}
	profileCache := &profileCacheStub{}
	service := NewProfileService(accountRepo, &videoStatsReaderStub{}, &followStatsReaderStub{}, profileCache)

	if _, err := service.GetProfile(t.Context(), 99); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("GetProfile() error = %v, want ErrAccountNotFound", err)
	}
	if profileCache.setCalls != 0 {
		t.Fatalf("missing account cache set calls = %d, want 0", profileCache.setCalls)
	}
}
