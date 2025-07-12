package main

import (
	"context"
	"discord-gemini-bot/src/agent"
	"discord-gemini-bot/src/discordbot"
	"discord-gemini-bot/src/models"
	"discord-gemini-bot/src/tools"
	"discord-gemini-bot/src/types"
	"discord-gemini-bot/src/utils"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Configuration constants
const (
	MEMORY_WINDOW_SIZE         = 20
	DISCORD_MAX_MESSAGE_LENGTH = 2000
)

// Global variables
var (
	discordBotToken     string
	geminiAPIKey        string
	model               models.LLMModel
	toolList            []tools.Tool
	channelAgents       map[string]*agent.Agent
	supportedImageTypes = []string{"image/png", "image/jpeg", "image/webp", "image/gif"}
)

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Get environment variables
	discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	geminiAPIKey = os.Getenv("GEMINI_API_KEY")

	if discordBotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Initialize model
	var err error
	model, err = models.NewGemini(geminiAPIKey, "gemini-2.0-flash-exp")
	if err != nil {
		log.Fatalf("Failed to initialize Gemini model: %v", err)
	}

	// Initialize tools
	toolList = []tools.Tool{
		tools.NewGoogleSearchTool(),
		tools.NewURLFetchTool(),
	}

	// Initialize channel agents map
	channelAgents = make(map[string]*agent.Agent)
}

func main() {
	bot, err := discordbot.NewBot(discordBotToken, messageHandler)
	if err != nil {
		log.Fatalf("Failed to create Discord bot: %v", err)
	}
	if err := bot.Run(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
	if geminiModel, ok := model.(*models.Gemini); ok {
		geminiModel.Close()
	}
}

// messageHandler handles Discord messages and is passed to the bot
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages sent by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the bot is mentioned
	isBotMentioned := false
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			isBotMentioned = true
			break
		}
	}

	if !isBotMentioned {
		return
	}

	log.Printf("Received message from %s in channel %s: %s", m.Author.Username, m.ChannelID, m.Content)

	// Convert Discord message to internal Message type
	msg := types.DiscordMessageToMessage(s, m, supportedImageTypes)

	// Start typing indicator
	err := s.ChannelTyping(m.ChannelID)
	if err != nil {
		log.Printf("Error starting typing indicator: %v", err)
	}

	// Get or create agent for this channel
	channelID := m.ChannelID
	currentAgent, exists := channelAgents[channelID]
	if !exists {
		log.Printf("Creating new agent for channel %s", channelID)
		memory := types.NewConversationMemory(MEMORY_WINDOW_SIZE)
		currentAgent = agent.NewAgent(model, memory, toolList)
		channelAgents[channelID] = currentAgent
	}

	// Add message to memory
	currentAgent.AddMessage(msg)

	// Get response from agent
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	responseText, err := currentAgent.GetResponse(ctx)
	if err != nil {
		log.Printf("Error getting response from agent: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Sorry! Something went wrong while processing your request. Please try again later.")
		return
	}

	// Handle Discord's message length limit
	if len(responseText) > DISCORD_MAX_MESSAGE_LENGTH {
		chunks := utils.SplitLongText(responseText, DISCORD_MAX_MESSAGE_LENGTH)
		for i, chunk := range chunks {
			_, err := s.ChannelMessageSend(m.ChannelID, chunk)
			if err != nil {
				log.Printf("Error sending message chunk %d: %v", i+1, err)
				break
			}
			// Add small delay between messages to avoid rate limits
			if i < len(chunks)-1 {
				time.Sleep(500 * time.Millisecond)
			}
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, responseText)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}

	log.Printf("Sent response to channel %s", m.ChannelID)
}
