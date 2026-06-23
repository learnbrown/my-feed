package video

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"my_feed/internal/db"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

type VideoService struct {
	repo *VideoRepo
}

func NewVideoService(repo *VideoRepo) *VideoService {
	return &VideoService{repo: repo}
}

const (
	maxVideoSize    int64  = 200 << 20 // 视频最大字节数
	maxCoverSize    int64  = 5 << 20   // 封面最大字节数
	defaultCoverURL string = "/static/uploads/covers/default.png"

	// 用于saveFile区分视频与封面
	videoType string = "videos"
	coverType string = "covers"
)

var (
	ErrVideoNotFound       = errors.New("video not found")
	ErrAuthorNotFound      = errors.New("author not found")
	ErrAuthorRequired      = errors.New("author required")
	ErrTitleRequired       = errors.New("title required")
	ErrPlayURLRequired     = errors.New("play url required")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrFileTooLarge        = errors.New("file too large")
	ErrFileRequired        = errors.New("file required")
	ErrInvalidPlayURL      = errors.New("invalid play url")
	ErrInvalidCoverURL     = errors.New("invalid cover url")
)

var tagRegex = regexp.MustCompile(`#[\p{L}\p{N}_+]+`)

func (service *VideoService) Publish(authorID uint, title, description, video_url, cover_url string) (*Video, error) {
	// [x] 缺少业务校验
	if authorID == 0 {
		return nil, ErrAuthorRequired
	}
	// 去除首尾空白字符
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrTitleRequired
	}
	video_url = strings.TrimSpace(video_url)
	if video_url == "" {
		return nil, ErrPlayURLRequired
	}
	if !strings.HasPrefix(video_url, "/static/uploads/videos/") {
		return nil, ErrInvalidPlayURL
	}

	cover_url = strings.TrimSpace(cover_url)
	if cover_url == "" {
		cover_url = defaultCoverURL
	}
	if !strings.HasPrefix(cover_url, "/static/uploads/covers/") {
		return nil, ErrInvalidCoverURL
	}

	description = strings.TrimSpace(description)

	video := &Video{
		AuthorID:    authorID,
		Title:       title,
		Description: description,
		PlayURL:     video_url,
		CoverURL:    cover_url,
	}

	// 标签提取
	tags := tagRegex.FindAllString(title+" "+description, -1)
	cleanTags := []string{}
	// 去除tag前的#并转小写
	for _, value := range tags {
		value = strings.TrimPrefix(value, "#")
		value = strings.TrimSpace(value)
		value = strings.ToLower(value)
		if value == "" {
			continue
		}
		cleanTags = append(cleanTags, value)
	}
	// 去重
	slices.Sort(cleanTags)
	uniqueTags := slices.Compact(cleanTags)

	// 整合video, tag, video-tag创建逻辑，中间流程出错直接返回并回滚数据库事务，只返回一个error
	err := service.repo.Transaction(func(txRepo *VideoRepo) error {
		err := txRepo.CreateVideo(video)
		if err != nil {
			return err
		}

		for _, name := range uniqueTags {

			tag, err := txRepo.FindOrCreateTag(name)
			if err != nil {
				return err
			}

			err = txRepo.CreateVideoTag(video.ID, tag.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return video, err
}

func (service *VideoService) GetDetail(id uint) (*Video, error) {
	video, err := service.repo.FindVideoByID(id)
	if errors.Is(err, db.ErrRecordNotFound) {
		return nil, ErrVideoNotFound
	}
	return video, err
}

// 向handler返回list的结构
type ListResponse struct {
	Videos   []Video `json:"videos"`
	NextTime int64   `json:"next_time"`
	HasMore  bool    `json:"has_more"`
}

func (service *VideoService) ListByAuthorID(authorID uint, limit int, latestTime time.Time) (*ListResponse, error) {
	if authorID == 0 {
		return nil, ErrAuthorRequired
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	videos, err := service.repo.ListByAuthorID(authorID, limit+1, latestTime)
	if err != nil {
		return nil, err
	}

	// [x] 更严谨的hasMore判断
	// 查询limit+1条记录，如果查到了，说明还有下一页
	// 并取limit条记录返回
	hasMore := len(*videos) > limit
	if hasMore {
		*videos = (*videos)[:limit]
	}

	var nextTime int64
	if len(*videos) > 0 {
		nextTime = (*videos)[len(*videos)-1].CreatedAt.UnixMilli()
	}

	res := &ListResponse{
		Videos:   *videos,
		NextTime: nextTime,
		HasMore:  hasMore,
	}

	return res, err
}

func (service *VideoService) UploadVideo(authorID uint, file *multipart.FileHeader) (string, error) {
	return service.saveFile(authorID, videoType, file)
}

func (service *VideoService) UploadCover(authorID uint, file *multipart.FileHeader) (string, error) {
	return service.saveFile(authorID, coverType, file)
}

func (service *VideoService) saveFile(authorID uint, fileType string, file *multipart.FileHeader) (string, error) {
	if authorID == 0 {
		return "", ErrAuthorRequired
	}
	if file == nil {
		return "", ErrFileRequired
	}

	if fileType != videoType && fileType != coverType {
		// TODO: 先返回此种错误
		return "", ErrUnsupportedFileType
	}

	// TODO: 改为依靠文件头判断类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if fileType == videoType && ext != ".mp4" {
		return "", ErrUnsupportedFileType
	} else if fileType == coverType && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return "", ErrUnsupportedFileType
	}

	// 限制文件大小
	if fileType == videoType && file.Size > maxVideoSize {
		return "", ErrFileTooLarge
	} else if fileType == coverType && file.Size > maxCoverSize {
		return "", ErrFileTooLarge
	}

	// 本地文件路径 .run/uploads/<fileType>/<authorID>/<yyyyMMdd>/<uniqueName><ext>
	dateDir := time.Now().Format("20060102")

	uploadDir := filepath.Join(".run", "uploads", fileType, fmt.Sprintf("%d", authorID), dateDir)

	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return "", err
	}

	// 唯一文件名：时间戳纳秒+随机数
	randomNum, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return "", err
	}
	uniqueName := fmt.Sprintf("%d_%05d", time.Now().UnixNano(), randomNum.Int64())

	fileName := uniqueName + ext

	filePath := filepath.Join(uploadDir, fileName)

	// 打开文件
	srcFile, err := file.Open()
	if err != nil {
		return "", err
	}
	// 结束时关闭文件
	defer srcFile.Close()

	// 创建目标文件
	// os.O_CREATE: 不存在则创建 | os.O_WRONLY: 只写模式 | os.O_TRUNC: 存在则清空覆盖
	dstFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	// 结束时关闭文件
	defer dstFile.Close()

	// io.Copy 会在底层建立一块缓冲区，高效地把数据从 srcFile 读出并写入 dstFile，直到遇到 EOF（文件末尾）
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", err
	}

	returnDir := fmt.Sprintf("/static/uploads/%s/%d/%s/%s", fileType, authorID, dateDir, fileName)

	return returnDir, nil
}
