package main

import (
	"context"
	"discord-gemini-bot/src/models"
	"discord-gemini-bot/src/types"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestGeminiModel(t *testing.T) {
	// Load environment variables for testing
	_ = godotenv.Load("../.env")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping Gemini tests")
	}

	// Test model creation
	t.Run("CreateModel", func(t *testing.T) {
		model, err := models.NewGemini(apiKey, "gemini-2.0-flash-exp")
		if err != nil {
			t.Fatalf("Failed to create Gemini model: %v", err)
		}

		if model == nil {
			t.Fatal("Model is nil")
		}

		defer model.Close()
	})

	// Test with empty API key
	t.Run("CreateModelWithEmptyAPIKey", func(t *testing.T) {
		_, err := models.NewGemini("", "gemini-2.0-flash-exp")
		if err == nil {
			t.Fatal("Expected error with empty API key")
		}
	})

	// Test system prompt setting
	t.Run("SetSystemPrompt", func(t *testing.T) {
		model, err := models.NewGemini(apiKey, "gemini-2.0-flash-exp")
		if err != nil {
			t.Fatalf("Failed to create Gemini model: %v", err)
		}
		defer model.Close()

		systemPrompt := "You are a helpful test assistant."
		model.SetSystemPrompt(systemPrompt)

		// Note: We can't directly test the private field, but we can test that no error occurs
	})

	// Test model info
	t.Run("GetModelInfo", func(t *testing.T) {
		model, err := models.NewGemini(apiKey, "gemini-2.0-flash-exp")
		if err != nil {
			t.Fatalf("Failed to create Gemini model: %v", err)
		}
		defer model.Close()

		info := model.GetModelInfo()
		if info == nil {
			t.Fatal("Model info is nil")
		}

		if name, exists := info["name"]; !exists || name != "gemini-2.0-flash-exp" {
			t.Errorf("Expected model name to be 'gemini-2.0-flash-exp', got %v", name)
		}
	})
}

// Integration tests that require actual API calls
func TestGeminiIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Load environment variables for testing
	_ = godotenv.Load("../.env")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration tests")
	}

	model, err := models.NewGemini(apiKey, "gemini-2.0-flash-exp")
	if err != nil {
		t.Fatalf("Failed to create Gemini model: %v", err)
	}
	defer model.Close()

	// Test simple generation
	t.Run("SimpleGeneration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := model.GenerateAsync(ctx, "What is 2+2? Answer with just the number.", nil)
		if err != nil {
			t.Fatalf("Failed to generate response: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from model")
		}

		t.Logf("Model response: %s", response)
	})

	// Test generation with history
	t.Run("GenerationWithHistory", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		messages := []*types.Message{
			types.NewMessage("user", "text", "My name is Alice."),
			types.NewMessage("assistant", "text", "Hello Alice! Nice to meet you."),
			types.NewMessage("user", "text", "What is my name?"),
		}

		response, err := model.GenerateWithHistoryAsync(ctx, messages)
		if err != nil {
			t.Fatalf("Failed to generate response with history: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from model")
		}

		t.Logf("Model response with history: %s", response)
	})

	// Test with system prompt
	t.Run("GenerationWithSystemPrompt", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		model.SetSystemPrompt("You are a math tutor. Always show your work.")

		response, err := model.GenerateAsync(ctx, "What is 15 + 27?", nil)
		if err != nil {
			t.Fatalf("Failed to generate response: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from model")
		}

		t.Logf("Model response with system prompt: %s", response)
	})
}
