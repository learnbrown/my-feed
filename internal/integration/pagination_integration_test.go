package integration

import (
	"testing"
	"time"

	"my_feed/internal/account"
	"my_feed/internal/comment"
	dbtest "my_feed/internal/db/testutil"
	"my_feed/internal/feed"
	"my_feed/internal/follow"
	"my_feed/internal/like"
	"my_feed/internal/message"
	"my_feed/internal/video"

	"gorm.io/gorm"
)

func TestCompositeCursorPaginationWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)

	t.Run("video author, latest feed, and tag feed", func(t *testing.T) {
		dbtest.CleanTestDB(t, sqlDB)
		testVideoFeedPagination(t, sqlDB)
	})
	t.Run("liked videos", func(t *testing.T) {
		dbtest.CleanTestDB(t, sqlDB)
		testLikedVideoPagination(t, sqlDB)
	})
	t.Run("comments", func(t *testing.T) {
		dbtest.CleanTestDB(t, sqlDB)
		testCommentPagination(t, sqlDB)
	})
	t.Run("followers and followings", func(t *testing.T) {
		dbtest.CleanTestDB(t, sqlDB)
		testFollowPagination(t, sqlDB)
	})
	t.Run("conversation", func(t *testing.T) {
		dbtest.CleanTestDB(t, sqlDB)
		testConversationPagination(t, sqlDB)
	})
}

func testVideoFeedPagination(t *testing.T, sqlDB *gorm.DB) {
	author := createTestAccount(t, sqlDB, "pagination-author")
	createdAt := fixedCursorTime()
	var visibleIDs []uint
	var videos []video.Video
	for i := 1; i <= 5; i++ {
		v := createTestVideo(t, sqlDB, author.ID, i, 1, createdAt)
		videos = append(videos, v)
		visibleIDs = append(visibleIDs, v.ID)
	}
	hidden := createTestVideo(t, sqlDB, author.ID, 99, 2, createdAt)

	tag := video.Tag{Name: "go"}
	if err := sqlDB.Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}
	for _, v := range append(videos, hidden) {
		if err := sqlDB.Create(&video.VideoTag{VideoID: v.ID, TagID: tag.ID}).Error; err != nil {
			t.Fatalf("create video tag: %v", err)
		}
	}

	videoRepo := video.NewVideoRepo(sqlDB)
	videoService := video.NewVideoService(videoRepo, nil)
	feedService := feed.NewFeedService(videoRepo)
	want := reverseUintIDs(visibleIDs)

	authorIDs := collectVideoPages(t, func(latestTime time.Time, latestID uint) (*video.ListResponse, error) {
		return videoService.ListByAuthorID(author.ID, 2, latestTime, latestID)
	})
	assertUintIDs(t, authorIDs, want)

	latestIDs := collectVideoPages(t, func(latestTime time.Time, latestID uint) (*video.ListResponse, error) {
		return feedService.ListLatest(2, latestTime, latestID)
	})
	assertUintIDs(t, latestIDs, want)

	tagIDs := collectVideoPages(t, func(latestTime time.Time, latestID uint) (*video.ListResponse, error) {
		return feedService.ListByTag("#GO", 2, latestTime, latestID)
	})
	assertUintIDs(t, tagIDs, want)
}

func collectVideoPages(t *testing.T, fetch func(time.Time, uint) (*video.ListResponse, error)) []uint {
	t.Helper()

	var ids []uint
	var latestTime time.Time
	var latestID uint
	for page := 0; page < 10; page++ {
		response, err := fetch(latestTime, latestID)
		if err != nil {
			t.Fatalf("fetch video page %d: %v", page, err)
		}
		for _, item := range response.Videos {
			ids = append(ids, item.ID)
		}
		if !response.HasMore {
			return ids
		}
		if response.NextTime == 0 || response.NextID == 0 {
			t.Fatalf("page %d has_more without next cursor: %#v", page, response)
		}
		latestTime = time.UnixMilli(response.NextTime)
		latestID = response.NextID
	}
	t.Fatal("video pagination did not terminate")
	return nil
}

func testLikedVideoPagination(t *testing.T, sqlDB *gorm.DB) {
	author := createTestAccount(t, sqlDB, "liked-author")
	liker := createTestAccount(t, sqlDB, "liked-user")
	createdAt := fixedCursorTime()
	var relationIDs []uint
	videoByRelation := make(map[uint]uint)
	for i := 1; i <= 5; i++ {
		v := createTestVideo(t, sqlDB, author.ID, i, 1, createdAt.Add(-time.Duration(i)*time.Hour))
		relation := like.Like{AccountID: liker.ID, VideoID: v.ID, CreatedAt: createdAt}
		if err := sqlDB.Create(&relation).Error; err != nil {
			t.Fatalf("create like: %v", err)
		}
		relationIDs = append(relationIDs, relation.ID)
		videoByRelation[relation.ID] = v.ID
	}

	want := make([]uint, 0, len(relationIDs))
	for _, relationID := range reverseUintIDs(relationIDs) {
		want = append(want, videoByRelation[relationID])
	}
	service := like.NewLikeService(like.NewLikeRepo(sqlDB), nil)
	var got []uint
	var latestTime time.Time
	var latestID uint
	for page := 0; page < 10; page++ {
		response, err := service.ListLikedVideos(liker.ID, 2, latestTime, latestID)
		if err != nil {
			t.Fatalf("list liked page %d: %v", page, err)
		}
		for _, item := range response.Likes {
			got = append(got, item.ID)
		}
		if !response.HasMore {
			break
		}
		latestTime = time.UnixMilli(response.NextTime)
		latestID = response.NextID
	}
	assertUintIDs(t, got, want)
}

func testCommentPagination(t *testing.T, sqlDB *gorm.DB) {
	author := createTestAccount(t, sqlDB, "comment-author")
	commenter := createTestAccount(t, sqlDB, "comment-user")
	v := createTestVideo(t, sqlDB, author.ID, 1, 1, fixedCursorTime().Add(-time.Hour))
	createdAt := fixedCursorTime()
	var ids []uint
	for i := 1; i <= 5; i++ {
		item := comment.Comment{VideoID: v.ID, AccountID: commenter.ID, Content: "comment", CreatedAt: createdAt}
		if err := sqlDB.Create(&item).Error; err != nil {
			t.Fatalf("create comment: %v", err)
		}
		ids = append(ids, item.ID)
	}
	deleted := comment.Comment{VideoID: v.ID, AccountID: commenter.ID, Content: "deleted", CreatedAt: createdAt}
	if err := sqlDB.Create(&deleted).Error; err != nil {
		t.Fatalf("create deleted comment: %v", err)
	}
	if err := sqlDB.Delete(&deleted).Error; err != nil {
		t.Fatalf("soft delete comment: %v", err)
	}

	service := comment.NewCommentService(comment.NewCommentRepo(sqlDB), nil)
	var got []uint
	var latestTime time.Time
	var latestID uint
	for page := 0; page < 10; page++ {
		response, err := service.ListComment(v.ID, 2, latestTime, latestID)
		if err != nil {
			t.Fatalf("list comments page %d: %v", page, err)
		}
		for _, item := range response.Comments {
			got = append(got, item.ID)
		}
		if !response.HasMore {
			break
		}
		latestTime = time.UnixMilli(response.NextTime)
		latestID = response.NextID
	}
	assertUintIDs(t, got, reverseUintIDs(ids))
}

func testFollowPagination(t *testing.T, sqlDB *gorm.DB) {
	owner := createTestAccount(t, sqlDB, "follow-owner")
	createdAt := fixedCursorTime()
	var followerRelationIDs []uint
	followerByRelation := make(map[uint]uint)
	var followingRelationIDs []uint
	followingByRelation := make(map[uint]uint)
	for i := 1; i <= 5; i++ {
		follower := createTestAccount(t, sqlDB, "follower-"+string(rune('a'+i)))
		followerRelation := follow.Follow{FollowerID: follower.ID, VloggerID: owner.ID, CreatedAt: createdAt}
		if err := sqlDB.Create(&followerRelation).Error; err != nil {
			t.Fatalf("create follower relation: %v", err)
		}
		followerRelationIDs = append(followerRelationIDs, followerRelation.ID)
		followerByRelation[followerRelation.ID] = follower.ID

		vlogger := createTestAccount(t, sqlDB, "vlogger-"+string(rune('a'+i)))
		followingRelation := follow.Follow{FollowerID: owner.ID, VloggerID: vlogger.ID, CreatedAt: createdAt}
		if err := sqlDB.Create(&followingRelation).Error; err != nil {
			t.Fatalf("create following relation: %v", err)
		}
		followingRelationIDs = append(followingRelationIDs, followingRelation.ID)
		followingByRelation[followingRelation.ID] = vlogger.ID
	}

	service := follow.NewFollowService(follow.NewFollowRepo(sqlDB), accountRepo(sqlDB))
	followerIDs := collectFollowPages(t, func(latestTime time.Time, latestID uint) (*follow.ListFollowResponse, error) {
		return service.ListFollower(owner.ID, 2, latestTime, latestID)
	})
	wantFollowers := make([]uint, 0, len(followerRelationIDs))
	for _, relationID := range reverseUintIDs(followerRelationIDs) {
		wantFollowers = append(wantFollowers, followerByRelation[relationID])
	}
	assertUintIDs(t, followerIDs, wantFollowers)

	followingIDs := collectFollowPages(t, func(latestTime time.Time, latestID uint) (*follow.ListFollowResponse, error) {
		return service.ListFollowing(owner.ID, 2, latestTime, latestID)
	})
	wantFollowings := make([]uint, 0, len(followingRelationIDs))
	for _, relationID := range reverseUintIDs(followingRelationIDs) {
		wantFollowings = append(wantFollowings, followingByRelation[relationID])
	}
	assertUintIDs(t, followingIDs, wantFollowings)
}

func collectFollowPages(t *testing.T, fetch func(time.Time, uint) (*follow.ListFollowResponse, error)) []uint {
	t.Helper()

	var ids []uint
	var latestTime time.Time
	var latestID uint
	for page := 0; page < 10; page++ {
		response, err := fetch(latestTime, latestID)
		if err != nil {
			t.Fatalf("fetch follow page %d: %v", page, err)
		}
		for _, item := range response.Follows {
			ids = append(ids, item.ID)
		}
		if !response.HasMore {
			return ids
		}
		latestTime = time.UnixMilli(response.NextTime)
		latestID = response.NextID
	}
	t.Fatal("follow pagination did not terminate")
	return nil
}

func testConversationPagination(t *testing.T, sqlDB *gorm.DB) {
	alice := createTestAccount(t, sqlDB, "message-alice")
	bob := createTestAccount(t, sqlDB, "message-bob")
	outsider := createTestAccount(t, sqlDB, "message-outsider")
	createdAt := fixedCursorTime()
	var ids []uint
	for i := 1; i <= 5; i++ {
		item := message.Message{FromID: alice.ID, ToID: bob.ID, Content: "message", CreatedAt: createdAt}
		if i%2 == 0 {
			item.FromID, item.ToID = item.ToID, item.FromID
		}
		if err := sqlDB.Create(&item).Error; err != nil {
			t.Fatalf("create message: %v", err)
		}
		ids = append(ids, item.ID)
	}
	if err := sqlDB.Create(&message.Message{FromID: alice.ID, ToID: outsider.ID, Content: "outside", CreatedAt: createdAt}).Error; err != nil {
		t.Fatalf("create outsider message: %v", err)
	}

	service := message.NewMessageService(message.NewMessageRepo(sqlDB), accountRepo(sqlDB))
	var got []uint
	var latestTime time.Time
	var latestID uint
	for page := 0; page < 10; page++ {
		response, err := service.ListConversation(alice.ID, bob.ID, 2, latestTime, latestID)
		if err != nil {
			t.Fatalf("list conversation page %d: %v", page, err)
		}
		for _, item := range response.Messages {
			got = append(got, item.ID)
		}
		if !response.HasMore {
			break
		}
		latestTime = time.UnixMilli(response.NextTime)
		latestID = response.NextID
	}
	assertUintIDs(t, got, reverseUintIDs(ids))
}

func accountRepo(sqlDB *gorm.DB) *account.AccountRepo {
	return account.NewAccountRepo(sqlDB)
}
