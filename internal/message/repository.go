package message

import (
	"time"

	"gorm.io/gorm"
)

type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (repo *MessageRepo) SendMessage(fromID, toID uint, content string) (*Message, error) {
	message := &Message{
		FromID:  fromID,
		ToID:    toID,
		Content: content,
	}
	err := repo.db.Create(message).Error
	return message, err
}

// TODO: 升级复合游标
func (repo *MessageRepo) ListConversation(fromID, toID uint, limit int, latestTime time.Time) (*[]Message, error) {
	// fromID: current ID toID: with account ID
	messages := &[]Message{}
	query := repo.db.Model(&Message{}).
		Where(
			"(from_id = ? AND to_id = ?) OR (to_id = ? AND from_id = ?)",
			fromID, toID, // 当前用户发给对方的消息
			fromID, toID, // 对方发给当前用户的消息
		)

	if !latestTime.IsZero() {
		query = query.Where("created_at < ?", latestTime)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Find(messages).Error

	return messages, err
}
