package message

import (
	"errors"
	"my_feed/internal/account"
	"my_feed/internal/db"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrReceiverRequired = errors.New("receiver required")
	ErrSenderRequired   = errors.New("sender required")
	ErrSendToYourself   = errors.New("can't send message to yourself")
	ErrContentRequired  = errors.New("content required")
	ErrContentTooLarge  = errors.New("content too large")
	ErrAccountNotFound  = errors.New("account not found")
	ErrInvalidCursor    = errors.New("invalid cursor")
)

type MessageService struct {
	msgRepo     *MessageRepo
	accountRepo *account.AccountRepo
}

func NewMessageService(msgRepo *MessageRepo, accountRepo *account.AccountRepo) *MessageService {
	return &MessageService{msgRepo: msgRepo, accountRepo: accountRepo}
}

type MessageDTO struct {
	ID        uint   `json:"id"`
	FromID    uint   `json:"from_id"`
	ToID      uint   `json:"to_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func (service *MessageService) SendMessage(fromID, toID uint, content string) (*MessageDTO, error) {
	if fromID == 0 {
		return nil, ErrSenderRequired
	}
	if toID == 0 {
		return nil, ErrReceiverRequired
	}

	if fromID == toID {
		return nil, ErrSendToYourself
	}

	_, err := service.accountRepo.FindAccountByID(toID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrContentRequired
	}

	if utf8.RuneCountInString(content) > 1000 {
		return nil, ErrContentTooLarge
	}

	msg, err := service.msgRepo.SendMessage(fromID, toID, content)
	if err != nil {
		return nil, err
	}

	dto := &MessageDTO{
		ID:        msg.ID,
		FromID:    msg.FromID,
		ToID:      msg.ToID,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.UnixMilli(),
	}

	return dto, nil
}

// listConversation 响应格式
type ListCvsResponse struct {
	Messages []MessageDTO `json:"messages"`
	HasMore  bool         `json:"has_more"`
	NextTime int64        `json:"next_time"`
	NextID   uint         `json:"next_id"`
}

func (service *MessageService) ListConversation(fromID, toID uint, limit int, latestTime time.Time, latestID uint) (*ListCvsResponse, error) {
	if fromID == 0 {
		return nil, ErrSenderRequired
	}
	if toID == 0 {
		return nil, ErrReceiverRequired
	}

	if fromID == toID {
		return nil, ErrSendToYourself
	}

	_, err := service.accountRepo.FindAccountByID(toID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	if latestTime.IsZero() != (latestID == 0) {
		return nil, ErrInvalidCursor
	}

	messages, err := service.msgRepo.ListConversation(fromID, toID, limit+1, latestTime, latestID)
	if err != nil {
		return nil, err
	}

	hasMore := len(*messages) > limit
	if hasMore {
		*messages = (*messages)[:limit]
	}

	var nextTime int64
	var nextID uint
	if len(*messages) > 0 {
		nextTime = (*messages)[len(*messages)-1].CreatedAt.UnixMilli()
		nextID = (*messages)[len(*messages)-1].ID
	}

	dtos := make([]MessageDTO, len(*messages))
	for i, m := range *messages {
		dtos[i] = MessageDTO{
			ID:        m.ID,
			FromID:    m.FromID,
			ToID:      m.ToID,
			Content:   m.Content,
			CreatedAt: m.CreatedAt.UnixMilli(),
		}
	}

	return &ListCvsResponse{
		Messages: dtos,
		HasMore:  hasMore,
		NextTime: nextTime,
		NextID:   nextID,
	}, nil
}
