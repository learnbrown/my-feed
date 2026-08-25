package integration

import (
	"errors"
	"my_feed/internal/account"
	dbtest "my_feed/internal/db/testutil"
	"my_feed/internal/like"
	"my_feed/internal/video"
	"reflect"
	"testing"
)

type likeResult struct {
	count uint
	err   error
}

func TestLikeAndUnlikeIdempotentWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := account.Account{
		Username:     "author",
		PasswordHash: "hash",
	}

	if err := sqlDB.Create(&author).Error; err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	liker := account.Account{
		Username:     "liker",
		PasswordHash: "hash",
	}
	if err := sqlDB.Create(&liker).Error; err != nil {
		t.Fatalf("failed to create lliker: %v", err)
	}

	v := video.Video{
		AuthorID: author.ID,
		Title:    "test video",
		PlayURL:  "/static/uploads/videos/test/demo.mp4",
		CoverURL: "/static/uploads/covers/default.png",
		Status:   1,
	}
	if err := sqlDB.Create(&v).Error; err != nil {
		t.Fatalf("failed to create video: %v", err)
	}

	likeRepo := like.NewLikeRepo(sqlDB)
	cacheRecorder := &detailCacheRecorder{}
	profileRecorder := &profileCacheRecorder{}
	likeService := like.NewLikeService(likeRepo, cacheRecorder, profileRecorder)

	likesCount, err := likeService.Like(t.Context(), liker.ID, v.ID)
	if err != nil {
		t.Fatalf("first like failed: %v", err)
	}
	if likesCount != 1 {
		t.Fatalf("first like expected likes_count 1, got %d", likesCount)
	}

	likesCount, err = likeService.Like(t.Context(), liker.ID, v.ID)
	if err != nil {
		t.Fatalf("second like failed: %v", err)
	}
	if likesCount != 1 {
		t.Fatalf("second like expected likes_count still 1, got %d", likesCount)
	}

	var likeRows int64
	if err := sqlDB.Model(&like.Like{}).
		Where("account_id = ? AND video_id = ?", liker.ID, v.ID).
		Count(&likeRows).Error; err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if likeRows != 1 {
		t.Fatalf("expected exactly 1 like row, got %d", likeRows)
	}

	var gotVideo video.Video
	if err := sqlDB.First(&gotVideo).Error; err != nil {
		t.Fatalf("failed to find video: %v", err)
	}

	if gotVideo.LikesCount != 1 {
		t.Fatalf("expected video.likes_count 1, got %d", gotVideo.LikesCount)
	}
	if len(cacheRecorder.deletedVideoIDs) != 1 || cacheRecorder.deletedVideoIDs[0] != v.ID {
		t.Fatalf("cache deletions after idempotent likes = %v, want [%d]", cacheRecorder.deletedVideoIDs, v.ID)
	}
	if !reflect.DeepEqual(profileRecorder.deletedAccountIDs, []uint{author.ID, author.ID}) {
		t.Fatalf("profile cache deletions after idempotent likes = %v, want [%d %d]", profileRecorder.deletedAccountIDs, author.ID, author.ID)
	}

	likesCount, err = likeService.Unlike(t.Context(), liker.ID, v.ID)
	if err != nil {
		t.Fatalf("first unlike failed: %v", err)
	}
	if likesCount != 0 {
		t.Fatalf("first unlike expected likes_count 0, got %d", likesCount)
	}

	likesCount, err = likeService.Unlike(t.Context(), liker.ID, v.ID)
	if err != nil {
		t.Fatalf("second unlike failed: %v", err)
	}
	if likesCount != 0 {
		t.Fatalf("second unlike expected likes_count still 0, got %d", likesCount)
	}

	if err := sqlDB.Model(&like.Like{}).
		Where("account_id = ? AND video_id = ?", liker.ID, v.ID).
		Count(&likeRows).Error; err != nil {
		t.Fatalf("count likes after unlike: %v", err)
	}
	if likeRows != 0 {
		t.Fatalf("expected no like rows after unlike, got %d", likeRows)
	}

	if err := sqlDB.First(&gotVideo, v.ID).Error; err != nil {
		t.Fatalf("failed to reload video: %v", err)
	}
	if gotVideo.LikesCount != 0 {
		t.Fatalf("expected video.likes_count 0 after unlike, got %d", gotVideo.LikesCount)
	}
	if len(cacheRecorder.deletedVideoIDs) != 2 || cacheRecorder.deletedVideoIDs[1] != v.ID {
		t.Fatalf("cache deletions after idempotent unlikes = %v, want [%d %d]", cacheRecorder.deletedVideoIDs, v.ID, v.ID)
	}
	if !reflect.DeepEqual(profileRecorder.deletedAccountIDs, []uint{author.ID, author.ID, author.ID, author.ID}) {
		t.Fatalf("profile cache deletions after idempotent unlikes = %v", profileRecorder.deletedAccountIDs)
	}

	profileRecorder.err = errors.New("redis delete unavailable")
	if _, err := likeService.Unlike(t.Context(), liker.ID, v.ID); err != nil {
		t.Fatalf("idempotent unlike with profile cache failure error = %v", err)
	}
}

func TestLikeConcurrentIdempotent(t *testing.T) {
	// 1. 准备测试 DB、用户、视频
	// 2. 启动 20 个 goroutine 同时调用 Like(accountID, videoID)
	// 3. 等全部结束
	// 4. 查询 likes 表，应该只有 1 条
	// 5. 查询 videos.likes_count，应该只加 1
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := account.Account{
		Username:     "author",
		PasswordHash: "hash",
	}

	if err := sqlDB.Create(&author).Error; err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	liker := account.Account{
		Username:     "liker",
		PasswordHash: "hash",
	}
	if err := sqlDB.Create(&liker).Error; err != nil {
		t.Fatalf("failed to create lliker: %v", err)
	}

	v := video.Video{
		AuthorID: author.ID,
		Title:    "test video",
		PlayURL:  "/static/uploads/videos/test/demo.mp4",
		CoverURL: "/static/uploads/covers/default.png",
		Status:   1,
	}
	if err := sqlDB.Create(&v).Error; err != nil {
		t.Fatalf("failed to create video: %v", err)
	}

	likeRepo := like.NewLikeRepo(sqlDB)
	likeService := like.NewLikeService(likeRepo, nil, nil)

	concurrency := 20

	results := make(chan likeResult, concurrency)
	ctx := t.Context()
	for i := 0; i < concurrency; i++ {
		go func() {
			cnt, err := likeService.Like(ctx, liker.ID, v.ID)
			results <- likeResult{count: cnt, err: err}
		}()
	}

	for i := 0; i < concurrency; i++ {
		result := <-results
		if result.err != nil {
			t.Errorf("concurrent like returned error: %v", result.err)
		}
		if result.count != 1 {
			t.Errorf("expected likes_count 1, got %d", result.count)
		}
	}

	var likeRows int64
	if err := sqlDB.Model(&like.Like{}).
		Where("account_id = ? AND video_id = ?", liker.ID, v.ID).
		Count(&likeRows).Error; err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if likeRows != 1 {
		t.Fatalf("expected exactly 1 like row, got %d", likeRows)
	}

	var gotVideo video.Video
	if err := sqlDB.First(&gotVideo).Error; err != nil {
		t.Fatalf("failed to find video: %v", err)
	}

	if gotVideo.LikesCount != 1 {
		t.Fatalf("expected video.likes_count 1, got %d", gotVideo.LikesCount)
	}

	results = make(chan likeResult, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			cnt, err := likeService.Unlike(ctx, liker.ID, v.ID)
			results <- likeResult{count: cnt, err: err}
		}()
	}
	for i := 0; i < concurrency; i++ {
		result := <-results
		if result.err != nil {
			t.Errorf("concurrent unlike returned error: %v", result.err)
		}
		if result.count != 0 {
			t.Errorf("expected likes_count 0, got %d", result.count)
		}
	}

	if err := sqlDB.Model(&like.Like{}).
		Where("account_id = ? AND video_id = ?", liker.ID, v.ID).
		Count(&likeRows).Error; err != nil {
		t.Fatalf("count likes after concurrent unlike: %v", err)
	}
	if likeRows != 0 {
		t.Fatalf("expected no like rows after concurrent unlike, got %d", likeRows)
	}
	if err := sqlDB.First(&gotVideo, v.ID).Error; err != nil {
		t.Fatalf("failed to reload video after concurrent unlike: %v", err)
	}
	if gotVideo.LikesCount != 0 {
		t.Fatalf("expected video.likes_count 0 after concurrent unlike, got %d", gotVideo.LikesCount)
	}
}
