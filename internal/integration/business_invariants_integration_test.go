package integration

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"my_feed/internal/account"
	"my_feed/internal/comment"
	dbtest "my_feed/internal/db/testutil"
	"my_feed/internal/follow"
	"my_feed/internal/message"
	"my_feed/internal/profile"
	"my_feed/internal/video"

	"gorm.io/gorm"
)

func TestCommentLifecycleMaintainsCountAndPermissionsWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := createTestAccount(t, sqlDB, "comment-lifecycle-author")
	commenter := createTestAccount(t, sqlDB, "comment-lifecycle-user")
	other := createTestAccount(t, sqlDB, "comment-lifecycle-other")
	v := createTestVideo(t, sqlDB, author.ID, 1, 1, fixedCursorTime())
	cacheRecorder := &detailCacheRecorder{}
	service := comment.NewCommentService(comment.NewCommentRepo(sqlDB), cacheRecorder)

	created, count, err := service.CreateComment(t.Context(), commenter.ID, v.ID, "  hello  ")
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if created.Content != "hello" || count != 1 {
		t.Fatalf("unexpected created comment/count: %#v count=%d", created, count)
	}
	assertVideoCommentCount(t, sqlDB, v.ID, 1)

	if _, err := service.DeleteComment(t.Context(), other.ID, created.ID); !errors.Is(err, comment.ErrCommentNotFound) {
		t.Fatalf("delete another user's comment error = %v, want ErrCommentNotFound", err)
	}
	assertVideoCommentCount(t, sqlDB, v.ID, 1)

	count, err = service.DeleteComment(t.Context(), commenter.ID, created.ID)
	if err != nil {
		t.Fatalf("DeleteComment() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("comments_count = %d, want 0", count)
	}
	assertVideoCommentCount(t, sqlDB, v.ID, 0)

	var activeComments int64
	if err := sqlDB.Model(&comment.Comment{}).Where("id = ?", created.ID).Count(&activeComments).Error; err != nil {
		t.Fatalf("count active comments: %v", err)
	}
	if activeComments != 0 {
		t.Fatalf("active comments = %d, want 0", activeComments)
	}
	if len(cacheRecorder.deletedVideoIDs) != 2 || cacheRecorder.deletedVideoIDs[0] != v.ID || cacheRecorder.deletedVideoIDs[1] != v.ID {
		t.Fatalf("cache deletions = %v, want [%d %d]", cacheRecorder.deletedVideoIDs, v.ID, v.ID)
	}
}

func TestFollowIsIdempotentWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	follower := createTestAccount(t, sqlDB, "follow-idempotent-follower")
	vlogger := createTestAccount(t, sqlDB, "follow-idempotent-vlogger")
	profileRecorder := &profileCacheRecorder{}
	service := follow.NewFollowService(follow.NewFollowRepo(sqlDB), account.NewAccountRepo(sqlDB), profileRecorder)

	if err := service.Follow(t.Context(), follower.ID, vlogger.ID); err != nil {
		t.Fatalf("first Follow() error = %v", err)
	}
	if err := service.Follow(t.Context(), follower.ID, vlogger.ID); err != nil {
		t.Fatalf("second Follow() error = %v", err)
	}
	assertFollowRows(t, sqlDB, follower.ID, vlogger.ID, 1)

	if err := service.Unfollow(t.Context(), follower.ID, vlogger.ID); err != nil {
		t.Fatalf("first Unfollow() error = %v", err)
	}
	if err := service.Unfollow(t.Context(), follower.ID, vlogger.ID); err != nil {
		t.Fatalf("second Unfollow() error = %v", err)
	}
	assertFollowRows(t, sqlDB, follower.ID, vlogger.ID, 0)

	if err := service.Follow(t.Context(), follower.ID, vlogger.ID+10_000); !errors.Is(err, follow.ErrVloggerNotFound) {
		t.Fatalf("Follow() missing vlogger error = %v, want ErrVloggerNotFound", err)
	}
	wantDeletedAccountIDs := []uint{
		follower.ID, vlogger.ID,
		follower.ID, vlogger.ID,
		follower.ID, vlogger.ID,
		follower.ID, vlogger.ID,
	}
	if !reflect.DeepEqual(profileRecorder.deletedAccountIDs, wantDeletedAccountIDs) {
		t.Fatalf("profile cache deletions = %v, want %v", profileRecorder.deletedAccountIDs, wantDeletedAccountIDs)
	}

	profileRecorder.err = errors.New("redis delete unavailable")
	if err := service.Unfollow(t.Context(), follower.ID, vlogger.ID); err != nil {
		t.Fatalf("idempotent Unfollow() with profile cache failure error = %v", err)
	}
}

func TestMessageValidationAndPersistenceWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	sender := createTestAccount(t, sqlDB, "message-validation-sender")
	receiver := createTestAccount(t, sqlDB, "message-validation-receiver")
	service := message.NewMessageService(message.NewMessageRepo(sqlDB), account.NewAccountRepo(sqlDB))

	if _, err := service.SendMessage(sender.ID, receiver.ID, " \t\n"); !errors.Is(err, message.ErrContentRequired) {
		t.Fatalf("empty message error = %v, want ErrContentRequired", err)
	}
	if _, err := service.SendMessage(sender.ID, receiver.ID, strings.Repeat("界", 1001)); !errors.Is(err, message.ErrContentTooLarge) {
		t.Fatalf("oversized message error = %v, want ErrContentTooLarge", err)
	}

	created, err := service.SendMessage(sender.ID, receiver.ID, "  hello  ")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if created.Content != "hello" {
		t.Fatalf("stored content = %q, want trimmed content", created.Content)
	}

	var stored message.Message
	if err := sqlDB.First(&stored, created.ID).Error; err != nil {
		t.Fatalf("find stored message: %v", err)
	}
	if stored.FromID != sender.ID || stored.ToID != receiver.ID || stored.Content != "hello" {
		t.Fatalf("unexpected stored message: %#v", stored)
	}
}

func TestProfileAggregatesOnlyVisibleVideosWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	owner := createTestAccount(t, sqlDB, "profile-owner")
	fan := createTestAccount(t, sqlDB, "profile-fan")
	following := createTestAccount(t, sqlDB, "profile-following")
	first := createTestVideo(t, sqlDB, owner.ID, 1, 1, fixedCursorTime())
	second := createTestVideo(t, sqlDB, owner.ID, 2, 1, fixedCursorTime())
	hidden := createTestVideo(t, sqlDB, owner.ID, 3, 2, fixedCursorTime())
	if err := sqlDB.Model(&video.Video{}).Where("id = ?", first.ID).UpdateColumn("likes_count", 3).Error; err != nil {
		t.Fatalf("set first likes count: %v", err)
	}
	if err := sqlDB.Model(&video.Video{}).Where("id = ?", second.ID).UpdateColumn("likes_count", 4).Error; err != nil {
		t.Fatalf("set second likes count: %v", err)
	}
	if err := sqlDB.Model(&video.Video{}).Where("id = ?", hidden.ID).UpdateColumn("likes_count", 100).Error; err != nil {
		t.Fatalf("set hidden likes count: %v", err)
	}
	if err := sqlDB.Create(&follow.Follow{FollowerID: fan.ID, VloggerID: owner.ID}).Error; err != nil {
		t.Fatalf("create fan relation: %v", err)
	}
	if err := sqlDB.Create(&follow.Follow{FollowerID: owner.ID, VloggerID: following.ID}).Error; err != nil {
		t.Fatalf("create following relation: %v", err)
	}

	service := profile.NewProfileService(account.NewAccountRepo(sqlDB), video.NewVideoRepo(sqlDB), follow.NewFollowRepo(sqlDB), nil)
	got, err := service.GetProfile(t.Context(), owner.ID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if got.Account.ID != owner.ID || got.Stats.VideosCount != 2 || got.Stats.LikesCount != 7 || got.Stats.FollowersCount != 1 || got.Stats.FollowingsCount != 1 {
		t.Fatalf("unexpected profile: %#v", got)
	}
}

func TestTransactionsRollBackAllWritesWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := createTestAccount(t, sqlDB, "rollback-author")
	rollbackErr := errors.New("force rollback")
	repo := video.NewVideoRepo(sqlDB)
	err := repo.Transaction(func(txRepo *video.VideoRepo) error {
		v := video.Video{
			AuthorID: author.ID,
			Title:    "rolled back",
			PlayURL:  "/static/uploads/videos/test/rollback.mp4",
			CoverURL: "/static/uploads/covers/default.png",
			Status:   1,
		}
		if err := txRepo.CreateVideo(&v); err != nil {
			return err
		}
		tag, err := txRepo.FindOrCreateTag("rollback")
		if err != nil {
			return err
		}
		if err := txRepo.CreateVideoTag(v.ID, tag.ID); err != nil {
			return err
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("Transaction() error = %v, want rollback error", err)
	}

	for name, model := range map[string]any{
		"videos":     &video.Video{},
		"tags":       &video.Tag{},
		"video_tags": &video.VideoTag{},
	} {
		var count int64
		if err := sqlDB.Model(model).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", name, err)
		}
		if count != 0 {
			t.Fatalf("%s count = %d after rollback, want 0", name, count)
		}
	}
}

func TestPublishNormalizesAndDeduplicatesTagsWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := createTestAccount(t, sqlDB, "publish-tags-author")
	profileRecorder := &profileCacheRecorder{}
	service := video.NewVideoService(video.NewVideoRepo(sqlDB), nil, profileRecorder)

	first, err := service.Publish(
		t.Context(),
		author.ID,
		"  Intro #Go #GO #C++  ",
		"  #go #feed #FEED #中文_1  ",
		"/static/uploads/videos/test/tags-1.mp4",
		"",
	)
	if err != nil {
		t.Fatalf("first Publish() error = %v", err)
	}
	if first.Title != "Intro #Go #GO #C++" || first.Description != "#go #feed #FEED #中文_1" || first.CoverURL != "/static/uploads/covers/default.png" {
		t.Fatalf("unexpected normalized video: %#v", first)
	}

	profileRecorder.err = errors.New("redis delete unavailable")
	second, err := service.Publish(
		t.Context(),
		author.ID,
		"Second #go",
		"reuse an existing tag",
		"/static/uploads/videos/test/tags-2.mp4",
		"/static/uploads/covers/test/tags-2.png",
	)
	if err != nil {
		t.Fatalf("second Publish() error = %v", err)
	}

	var tags []video.Tag
	if err := sqlDB.Order("name ASC").Find(&tags).Error; err != nil {
		t.Fatalf("list tags: %v", err)
	}
	wantTags := []string{"c++", "feed", "go", "中文_1"}
	if len(tags) != len(wantTags) {
		t.Fatalf("tag count = %d, want %d; tags=%#v", len(tags), len(wantTags), tags)
	}
	for i, want := range wantTags {
		if tags[i].Name != want {
			t.Fatalf("tags[%d] = %q, want %q", i, tags[i].Name, want)
		}
	}

	assertVideoTagRows(t, sqlDB, first.ID, 4)
	assertVideoTagRows(t, sqlDB, second.ID, 1)
	if !reflect.DeepEqual(profileRecorder.deletedAccountIDs, []uint{author.ID, author.ID}) {
		t.Fatalf("profile cache deletions after publish = %v, want [%d %d]", profileRecorder.deletedAccountIDs, author.ID, author.ID)
	}
}

func TestConcurrentCommentsMaintainCounterWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := createTestAccount(t, sqlDB, "concurrent-comments-author")
	commenter := createTestAccount(t, sqlDB, "concurrent-comments-user")
	v := createTestVideo(t, sqlDB, author.ID, 1, 1, fixedCursorTime())
	service := comment.NewCommentService(comment.NewCommentRepo(sqlDB), nil)

	const concurrency = 20
	start := make(chan struct{})
	errorsByRequest := make(chan error, concurrency)
	var ready sync.WaitGroup
	ready.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			ready.Done()
			<-start
			_, _, err := service.CreateComment(t.Context(), commenter.ID, v.ID, "concurrent comment")
			errorsByRequest <- err
		}()
	}
	ready.Wait()
	close(start)

	for i := 0; i < concurrency; i++ {
		if err := <-errorsByRequest; err != nil {
			t.Errorf("concurrent CreateComment() error = %v", err)
		}
	}

	var commentsCount int64
	if err := sqlDB.Model(&comment.Comment{}).Where("video_id = ?", v.ID).Count(&commentsCount).Error; err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if commentsCount != concurrency {
		t.Fatalf("comment rows = %d, want %d", commentsCount, concurrency)
	}
	assertVideoCommentCount(t, sqlDB, v.ID, concurrency)
}

func TestConcurrentFollowIsIdempotentWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	follower := createTestAccount(t, sqlDB, "concurrent-follow-follower")
	vlogger := createTestAccount(t, sqlDB, "concurrent-follow-vlogger")
	service := follow.NewFollowService(follow.NewFollowRepo(sqlDB), account.NewAccountRepo(sqlDB), nil)

	const concurrency = 20
	start := make(chan struct{})
	errorsByRequest := make(chan error, concurrency)
	var ready sync.WaitGroup
	ready.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			ready.Done()
			<-start
			errorsByRequest <- service.Follow(t.Context(), follower.ID, vlogger.ID)
		}()
	}
	ready.Wait()
	close(start)

	for i := 0; i < concurrency; i++ {
		if err := <-errorsByRequest; err != nil {
			t.Errorf("concurrent Follow() error = %v", err)
		}
	}
	assertFollowRows(t, sqlDB, follower.ID, vlogger.ID, 1)
}

func assertVideoCommentCount(t *testing.T, sqlDB *gorm.DB, videoID, want uint) {
	t.Helper()

	var v video.Video
	if err := sqlDB.First(&v, videoID).Error; err != nil {
		t.Fatalf("find video %d: %v", videoID, err)
	}
	if v.CommentsCount != want {
		t.Fatalf("video.comments_count = %d, want %d", v.CommentsCount, want)
	}
}

func assertFollowRows(t *testing.T, sqlDB *gorm.DB, followerID, vloggerID uint, want int64) {
	t.Helper()

	var count int64
	if err := sqlDB.Model(&follow.Follow{}).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Count(&count).Error; err != nil {
		t.Fatalf("count follow rows: %v", err)
	}
	if count != want {
		t.Fatalf("follow rows = %d, want %d", count, want)
	}
}

func assertVideoTagRows(t *testing.T, sqlDB *gorm.DB, videoID uint, want int64) {
	t.Helper()

	var count int64
	if err := sqlDB.Model(&video.VideoTag{}).Where("video_id = ?", videoID).Count(&count).Error; err != nil {
		t.Fatalf("count video tag rows: %v", err)
	}
	if count != want {
		t.Fatalf("video tag rows for video %d = %d, want %d", videoID, count, want)
	}
}
