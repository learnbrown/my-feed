package message

import (
	"errors"
	"testing"
	"time"
)

func TestSendMessageValidatesParticipants(t *testing.T) {
	service := MessageService{}

	tests := []struct {
		name    string
		fromID  uint
		toID    uint
		wantErr error
	}{
		{name: "sender required", toID: 1, wantErr: ErrSenderRequired},
		{name: "receiver required", fromID: 1, wantErr: ErrReceiverRequired},
		{name: "send to yourself", fromID: 1, toID: 1, wantErr: ErrSendToYourself},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.SendMessage(tt.fromID, tt.toID, "hello")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SendMessage() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListConversationValidatesParticipants(t *testing.T) {
	service := MessageService{}

	tests := []struct {
		name    string
		fromID  uint
		toID    uint
		wantErr error
	}{
		{name: "sender required", toID: 1, wantErr: ErrSenderRequired},
		{name: "receiver required", fromID: 1, wantErr: ErrReceiverRequired},
		{name: "conversation with yourself", fromID: 1, toID: 1, wantErr: ErrSendToYourself},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ListConversation(tt.fromID, tt.toID, 10, time.Time{}, 0)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListConversation() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
