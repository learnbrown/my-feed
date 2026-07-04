package integration

import (
	"my_feed/internal/account"
	dbtest "my_feed/internal/db/testutil"
	"my_feed/internal/like"
	"my_feed/internal/video"
	"testing"
)

func TestLikeIdempotentWithDB(t *testing.T) {
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
	likeService := like.NewLikeService(likeRepo, nil)

	likesCount, err := likeService.Like(nil, liker.ID, v.ID)
	if err != nil {
		t.Fatalf("first like failed: %v", err)
	}
	if likesCount != 1 {
		t.Fatalf("first like expected likes_count 1, got %d", likesCount)
	}

	likesCount, err = likeService.Like(nil, liker.ID, v.ID)
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
	likeService := like.NewLikeService(likeRepo, nil)

	concurrency := 20

	cntChan := make(chan uint, concurrency)
	errChan := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			cnt, err := likeService.Like(nil, liker.ID, v.ID)
			cntChan <- cnt
			errChan <- err
		}()
	}

	var errCount int
	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			errCount++
			t.Logf("concurrent like got error: %v", err)
		}
		if cnt := <-cntChan; cnt != 1 {
			t.Logf("expected likes_count 1, got:%d", cnt)
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
}
