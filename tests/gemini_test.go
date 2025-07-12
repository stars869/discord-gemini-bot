package tests

import (
	"context"
	"discord-gemini-bot/src/models"
	"os"
	"time"

	"testing"

	"github.com/joho/godotenv"
)

func TestGeminiSimpleGeneration(t *testing.T) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		t.Log("Warning: .env file not found")
	}

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Create Gemini model instance
	geminiModel, err := models.NewGemini(apiKey, "gemini-2.0-flash-exp")
	if err != nil {
		t.Fatalf("Failed to create Gemini model: %v", err)
	}

	// Set system prompt (if method exists)
	if setter, ok := interface{}(geminiModel).(interface{ SetSystemPrompt(string) }); ok {
		setter.SetSystemPrompt("You are a helpful AI assistant.")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate simple response
	t.Log("Testing simple generation:")
	response, err := geminiModel.GenerateAsync(ctx, "What is the capital of France?", nil)
	if err != nil {
		t.Errorf("Error in simple generation: %v", err)
	} else {
		t.Logf("Response: %s", response)
		if response == "" {
			t.Error("Expected a non-empty response from Gemini model")
		}
	}
	t.Log("Gemini model test completed successfully!")
}
