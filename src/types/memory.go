package types

import "container/list"

// ConversationMemory manages conversation history for a single channel
type ConversationMemory struct {
	windowSize int
	history    *list.List
}

// NewConversationMemory creates a new conversation memory with specified window size
func NewConversationMemory(windowSize int) *ConversationMemory {
	return &ConversationMemory{
		windowSize: windowSize,
		history:    list.New(),
	}
}

// AddMessage adds a message to the conversation history
func (cm *ConversationMemory) AddMessage(role, content string) {
	message := NewMessage(role, "text", content)
	cm.history.PushBack(message)
	
	// Remove oldest messages if we exceed window size
	for cm.history.Len() > cm.windowSize {
		cm.history.Remove(cm.history.Front())
	}
}

// GetHistory retrieves the conversation history as a slice of messages
func (cm *ConversationMemory) GetHistory() []*Message {
	messages := make([]*Message, 0, cm.history.Len())
	for e := cm.history.Front(); e != nil; e = e.Next() {
		messages = append(messages, e.Value.(*Message))
	}
	return messages
}

// Clear removes all messages from the conversation history
func (cm *ConversationMemory) Clear() {
	cm.history.Init()
}
