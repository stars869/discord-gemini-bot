package models

import (
	"context"
	"discord-gemini-bot/src/types"
)

// LLMModel is an abstract interface for Large Language Models
type LLMModel interface {
	// GenerateAsync generates text asynchronously based on the given prompt
	GenerateAsync(ctx context.Context, prompt string, images []map[string]interface{}) (string, error)
	
	// GenerateWithHistoryAsync generates text asynchronously with conversation history
	GenerateWithHistoryAsync(ctx context.Context, messages []*types.Message) (string, error)
	
	// SetSystemPrompt sets the system prompt for the model
	SetSystemPrompt(systemPrompt string)
}
