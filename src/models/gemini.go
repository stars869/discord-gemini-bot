package models

import (
	"context"
	"discord-gemini-bot/src/types"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Gemini implements the LLMModel interface using Google's Gemini API
type Gemini struct {
	client       *genai.Client
	model        *genai.GenerativeModel
	apiKey       string
	modelName    string
	systemPrompt string
	temperature  float32
	maxTokens    int32
}

// NewGemini creates a new Gemini model instance
func NewGemini(apiKey, modelName string) (*Gemini, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	
	if modelName == "" {
		modelName = "gemini-2.0-flash-exp"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(modelName)
	model.SetTemperature(1.0)
	model.SetMaxOutputTokens(8192)

	return &Gemini{
		client:      client,
		model:       model,
		apiKey:      apiKey,
		modelName:   modelName,
		temperature: 1.0,
		maxTokens:   8192,
	}, nil
}

// SetSystemPrompt sets the system prompt for the model
func (g *Gemini) SetSystemPrompt(systemPrompt string) {
	g.systemPrompt = systemPrompt
	g.model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}
}

// GenerateAsync generates text asynchronously based on the given prompt
func (g *Gemini) GenerateAsync(ctx context.Context, prompt string, images []map[string]interface{}) (string, error) {
	// Build parts for the request
	parts := []genai.Part{genai.Text(prompt)}
	
	// Add images if provided
	for _, image := range images {
		if data, ok := image["data"].(string); ok {
			if mimeType, ok := image["mime_type"].(string); ok {
				// Convert base64 string to blob
				blob := genai.Blob{
					MIMEType: mimeType,
					Data:     []byte(data), // This might need base64 decoding
				}
				parts = append(parts, blob)
			}
		}
	}

	resp, err := g.model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned")
	}

	// Extract text from the response
	var result string
	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

// GenerateWithHistoryAsync generates text asynchronously with conversation history
func (g *Gemini) GenerateWithHistoryAsync(ctx context.Context, messages []*types.Message) (string, error) {
	// Convert messages to genai.Content format
	contents := make([]*genai.Content, 0, len(messages))
	
	for _, msg := range messages {
		var role string
		switch msg.Role {
		case "user":
			role = "user"
		case "AI", "assistant":
			role = "model"
		default:
			role = "user"
		}
		
		content := &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		contents = append(contents, content)
	}

	// Create a chat session
	cs := g.model.StartChat()
	cs.History = contents

	// Generate response
	resp, err := cs.SendMessage(ctx, genai.Text("Please respond to the conversation"))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned")
	}

	// Extract text from the response
	var result string
	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

// Close closes the Gemini client
func (g *Gemini) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// GetModelInfo returns information about the model
func (g *Gemini) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        g.modelName,
		"temperature": g.temperature,
		"maxTokens":   g.maxTokens,
	}
}
