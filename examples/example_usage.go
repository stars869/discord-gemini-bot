package main

import (
	"context"
	"discord-gemini-bot/src/models"
	"discord-gemini-bot/src/types"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = "your-api-key-here"
		fmt.Println("Warning: Using placeholder API key. Set GEMINI_API_KEY environment variable.")
	}

	// Create Gemini model instance
	geminiModel, err := models.NewGemini(apiKey, "gemini-2.0-flash-exp")
	if err != nil {
		log.Fatalf("Failed to create Gemini model: %v", err)
	}
	defer geminiModel.Close()

	// Set system prompt
	geminiModel.SetSystemPrompt("You are a helpful AI assistant.")

	// Get model info
	fmt.Println("Model Info:")
	info := geminiModel.GetModelInfo()
	for key, value := range info {
		fmt.Printf("%s: %v\n", key, value)
	}
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate simple response
	fmt.Println("Simple Generation:")
	response, err := geminiModel.GenerateAsync(ctx, "What is the capital of France?", nil)
	if err != nil {
		log.Printf("Error in simple generation: %v", err)
	} else {
		fmt.Println(response)
	}
	fmt.Println()

	// Generate with conversation history
	fmt.Println("Generation with History:")
	messages := []*types.Message{
		types.NewMessage("user", "text", "Hello, what's your name?"),
		types.NewMessage("assistant", "text", "I'm Gemini, an AI assistant."),
		types.NewMessage("user", "text", "What can you help me with?"),
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	response, err = geminiModel.GenerateWithHistoryAsync(ctx2, messages)
	if err != nil {
		log.Printf("Error in generation with history: %v", err)
	} else {
		fmt.Println(response)
	}
}
