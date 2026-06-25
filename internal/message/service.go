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
)

type MessageService struct {
	msgRepo     *MessageRepo
	accountRepo *account.AccountRepo
}

func NewMessageService(msgRepo *MessageRepo, accountRepo *account.AccountRepo) *MessageService {
	return &MessageService{msgRepo: msgRepo, accountRepo: accountRepo}
}

func (service *MessageService) SendMessage(fromID, toID uint, content string) (*Message, error) {
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

	return service.msgRepo.SendMessage(fromID, toID, content)
}

// listConversation 响应格式
type ListCvsResponse struct {
	Messages *[]Message `json:"messages"`
	HasMore  bool       `json:"has_more"`
	NextTime int64      `json:"next_time"`
}

func (service *MessageService) ListConversation(fromID, toID uint, limit int, latestTime time.Time) (*ListCvsResponse, error) {
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

	messages, err := service.msgRepo.ListConversation(fromID, toID, limit+1, latestTime)
	if err != nil {
		return nil, err
	}

	hasMore := len(*messages) > limit
	if hasMore {
		*messages = (*messages)[:limit]
	}

	var nextTime int64
	if len(*messages) > 0 {
		nextTime = (*messages)[len(*messages)-1].CreatedAt.UnixMilli()
	}

	return &ListCvsResponse{
		Messages: messages,
		HasMore:  hasMore,
		NextTime: nextTime,
	}, nil
}
