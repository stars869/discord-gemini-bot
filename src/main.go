package main

import (
	"context"
	"discord-gemini-bot/src/agent"
	"discord-gemini-bot/src/models"
	"discord-gemini-bot/src/tools"
	"discord-gemini-bot/src/types"
	"discord-gemini-bot/src/utils"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Configuration constants
const (
	MEMORY_WINDOW_SIZE          = 20
	DISCORD_MAX_MESSAGE_LENGTH  = 2000
)

// Global variables
var (
	discordBotToken string
	geminiAPIKey    string
	model           models.LLMModel
	toolList        []tools.Tool
	channelAgents   map[string]*agent.Agent
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
	// Create a new Discord session
	dg, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Register event handlers
	dg.AddHandler(onReady)
	dg.AddHandler(onMessageCreate)

	// Set intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	// Open connection
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: %v", err)
	}

	// Wait for termination signal
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Close connections
	dg.Close()
	if geminiModel, ok := model.(*models.Gemini); ok {
		geminiModel.Close()
	}
}

// onReady is called when the bot successfully connects to Discord
func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v (ID: %v)", s.State.User.Username, s.State.User.Discriminator, s.State.User.ID)
	log.Println("Bot is ready!")
}

// onMessageCreate is called every time a message is sent in a channel the bot can see
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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

	// Remove the bot's mention from the message content
	cleanedContent := m.Content
	mentionString := fmt.Sprintf("<@%s>", s.State.User.ID)
	cleanedContent = strings.ReplaceAll(cleanedContent, mentionString, "")
	cleanedContent = strings.TrimSpace(cleanedContent)

	// Process images
	var images []map[string]interface{}
	for _, attachment := range m.Attachments {
		if attachment.ContentType != "" {
			for _, supportedType := range supportedImageTypes {
				if attachment.ContentType == supportedType {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					imageData, err := utils.GetImageAsBase64(ctx, attachment.URL)
					cancel()
					
					if err != nil {
						log.Printf("Error processing image: %v", err)
						continue
					}

					images = append(images, map[string]interface{}{
						"mime_type": imageData.MIMEType,
						"data":      imageData.Data,
					})
					break
				}
			}
		}
	}

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

	// Get response from agent
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	responseText, err := currentAgent.GetResponse(ctx, m.Author.Username, cleanedContent, images)
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
