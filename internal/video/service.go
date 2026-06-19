package video

type VideoService struct {
	repo *VideoRepo
}

func NewVideoService(repo *VideoRepo) *VideoService {
	return &VideoService{repo: repo}
}

// TODO: 这样好像会更新video的值，但是合适吗？
func (service *VideoService) Publish(title, description, video_url, cover_url string) (*Video, error) {
	video := &Video{
		Title:       title,
		Description: description,
		PlayURL:     video_url,
		CoverURL:    cover_url,
	}

	err := service.repo.CreateVideo(video)
	return video, err
}
