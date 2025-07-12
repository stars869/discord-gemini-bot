package types

import "time"

// Message represents a single message in the conversation
type Message struct {
	Timestamp time.Time `json:"timestamp"`
	Role      string    `json:"role"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
}

// NewMessage creates a new message with the current timestamp
func NewMessage(role, msgType, content string) *Message {
	return &Message{
		Timestamp: time.Now(),
		Role:      role,
		Type:      msgType,
		Content:   content,
	}
}
