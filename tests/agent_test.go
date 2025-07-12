package tests

import (
	"context"
	"discord-gemini-bot/src/agent"
	"discord-gemini-bot/src/models"
	"discord-gemini-bot/src/tools"
	"discord-gemini-bot/src/types"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestAgentResponse(t *testing.T) {
	t.Run("Setup", func(t *testing.T) {
		if err := godotenv.Load(); err != nil {
			t.Log("Warning: .env file not found")
		}
	})

	var geminiModel *models.Gemini
	var ag *agent.Agent

	t.Run("CreateModelAndAgent", func(t *testing.T) {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			t.Fatal("GEMINI_API_KEY environment variable is required")
		}

		var err error
		geminiModel, err = models.NewGemini(apiKey, "gemini-2.0-flash-exp")
		if err != nil {
			t.Fatalf("Failed to create Gemini model: %v", err)
		}
		// Create memory instance
		mem := &types.ConversationMemory{}
		ag = agent.NewAgent(geminiModel, mem, []tools.Tool{})
		if setter, ok := interface{}(ag).(interface{ SetSystemPrompt(string) }); ok {
			setter.SetSystemPrompt("You are a helpful AI agent.")
		}
	})

	t.Run("AgentResponse", func(t *testing.T) {
		defer geminiModel.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userMsg := types.NewMessage("user", []types.MessageContent{{Type: "text", Content: "Summarize the plot of Hamlet."}})
		ag.AddMessage(userMsg)

		t.Log("Testing agent response:")
		response, err := ag.GetResponse(ctx)
		if err != nil {
			t.Errorf("Error in agent response: %v", err)
		} else {
			t.Logf("Agent Response: %s", response)
			if response == "" {
				t.Error("Expected a non-empty response from Agent")
			}
		}
		t.Log("Agent test completed successfully!")
	})
}
