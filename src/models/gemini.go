package models

import (
	"context"
	"discord-gemini-bot/src/types"
	"encoding/base64"
	"fmt"

	"google.golang.org/genai"
)

// Gemini implements the LLMModel interface using Google's Gemini API
type Gemini struct {
	client       *genai.Client
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
	clientConfig := &genai.ClientConfig{
		APIKey: apiKey,
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &Gemini{
		client:      client,
		modelName:   modelName,
		temperature: 1.0,
		maxTokens:   8192,
	}, nil
}

// SetSystemPrompt sets the system prompt for the model
func (g *Gemini) SetSystemPrompt(systemPrompt string) {
	g.systemPrompt = systemPrompt
}

// GenerateAsync generates text asynchronously based on the given prompt
func (g *Gemini) GenerateAsync(ctx context.Context, prompt string, images []map[string]interface{}) (string, error) {
	// Build content parts
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}

	// Add images if provided
	if len(images) > 0 {
		parts := []*genai.Part{genai.NewPartFromText(prompt)}

		for _, image := range images {
			if data, ok := image["data"].(string); ok {
				if mimeType, ok := image["mime_type"].(string); ok {
					// Decode base64 data
					imageBytes, err := base64.StdEncoding.DecodeString(data)
					if err != nil {
						continue // Skip invalid base64 images
					}

					parts = append(parts, genai.NewPartFromBytes(imageBytes, mimeType))
				}
			}
		}

		contents = []*genai.Content{
			genai.NewContentFromParts(parts, genai.RoleUser),
		}
	}

	// Create generation config
	config := &genai.GenerateContentConfig{
		Temperature:     &g.temperature,
		MaxOutputTokens: g.maxTokens,
	}

	// Add system instruction if available
	if g.systemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(g.systemPrompt, genai.RoleUser)
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.modelName, contents, config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned")
	}

	// Extract text from the response
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			result += part.Text
		}
	}

	return result, nil
}

// GenerateWithHistoryAsync generates text asynchronously with conversation history
func (g *Gemini) GenerateWithHistoryAsync(ctx context.Context, messages []*types.Message) (string, error) {
	// Convert messages to genai.Content format using ToGenaiContent
	contents := make([]*genai.Content, 0, len(messages))
	for _, msg := range messages {
		content, err := msg.ToGenaiContent()
		if err != nil {
			return "", fmt.Errorf("failed to convert message to genai.Content: %w", err)
		}
		contents = append(contents, content)
	}

	// Create generation config
	config := &genai.GenerateContentConfig{
		Temperature:     &g.temperature,
		MaxOutputTokens: g.maxTokens,
	}

	// Add system instruction if available
	if g.systemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(g.systemPrompt, genai.RoleUser)
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.modelName, contents, config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content with history: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned")
	}

	// Extract text from the response
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			result += part.Text
		}
	}

	return result, nil
}

// Close closes the Gemini client (no-op for this implementation)
func (g *Gemini) Close() error {
	// The genai.Client doesn't have a Close method, so this is a no-op
	return nil
}
