package types

// ConversationMemory manages conversation history for a single channel
type ConversationMemory struct {
	windowSize int
	history    []*Message
}

// NewConversationMemory creates a new conversation memory with specified window size
func NewConversationMemory(windowSize int) *ConversationMemory {
	return &ConversationMemory{
		windowSize: windowSize,
		history:    make([]*Message, 0, windowSize),
	}
}

// AddMessage adds a message to the conversation history
func (cm *ConversationMemory) AddMessage(message *Message) {
	cm.history = append(cm.history, message)
	if len(cm.history) > cm.windowSize {
		cm.history = cm.history[1:]
	}
}

// GetHistory retrieves the conversation history as a slice of messages
func (cm *ConversationMemory) GetHistory() []*Message {
	return cm.history
}

// Clear removes all messages from the conversation history
func (cm *ConversationMemory) Clear() {
	cm.history = make([]*Message, 0, cm.windowSize)
}
