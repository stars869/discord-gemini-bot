package types

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/genai"
)

// Message represents a single message in the conversation
// MessageContent represents a single content item (text, image, etc.) in a message
type MessageContent struct {
	Type    string `json:"type"` // e.g., "text", "image"
	Content string `json:"content"`
}

// Message represents a single message in the conversation, which can have multiple contents
type Message struct {
	Timestamp time.Time        `json:"timestamp"`
	Role      string           `json:"role"`
	Contents  []MessageContent `json:"contents"`
}

// NewMessage creates a new message with the current timestamp and contents
func NewMessage(role string, contents []MessageContent) *Message {
	return &Message{
		Timestamp: time.Now(),
		Role:      role,
		Contents:  contents,
	}
}

// ToGenaiContent converts a Message to genai.Content format
func (m *Message) ToGenaiContent() (*genai.Content, error) {
	var role genai.Role
	switch m.Role {
	case "user":
		role = genai.RoleUser
	case "AI", "assistant":
		role = genai.RoleModel
	default:
		role = genai.RoleUser
	}

	// Convert all contents to []*genai.Part
	parts := make([]*genai.Part, 0, len(m.Contents))
	for _, c := range m.Contents {
		switch c.Type {
		case "text":
			parts = append(parts, &genai.Part{Text: c.Content})
		case "image":
			mimeType := "image/png" // default
			base64Data := c.Content
			if commaIdx := findMimeComma(c.Content); commaIdx > 0 {
				mimeType = c.Content[:commaIdx]
				base64Data = c.Content[commaIdx+1:]
			}
			imageBytes, err := decodeBase64(base64Data)
			if err != nil {
				return nil, err // return error if image decode fails
			}
			parts = append(parts, genai.NewPartFromBytes(imageBytes, mimeType))
		default:
			parts = append(parts, &genai.Part{Text: c.Content})
		}
	}

	return &genai.Content{
		Role:  string(role),
		Parts: parts,
	}, nil
}

// decodeBase64 decodes a base64 string and returns bytes
func decodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// findMimeComma finds the comma separating mime type and base64 data
func findMimeComma(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

// MessagesToGenaiContent converts a slice of Messages to a slice of genai.Content
func MessagesToGenaiContent(messages []*Message) ([]*genai.Content, error) {
	contents := make([]*genai.Content, 0, len(messages))
	for _, msg := range messages {
		content, err := msg.ToGenaiContent()
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	return contents, nil
}

// DiscordMessageToMessage converts a Discord message to our Message type
func DiscordMessageToMessage(s *discordgo.Session, m *discordgo.MessageCreate, supportedImageTypes []string) *Message {
	var contents []MessageContent

	// Remove bot mention from content
	cleanedContent := m.Content
	mentionString := "<@" + s.State.User.ID + ">"
	cleanedContent = strings.ReplaceAll(cleanedContent, mentionString, "")
	cleanedContent = strings.TrimSpace(cleanedContent)
	if cleanedContent != "" {
		contents = append(contents, MessageContent{Type: "text", Content: cleanedContent})
	}

	// Process images
	for _, attachment := range m.Attachments {
		if attachment.ContentType != "" {
			for _, supportedType := range supportedImageTypes {
				if attachment.ContentType == supportedType {
					// Assume image is already base64 encoded elsewhere, or fetch here if needed
					// For now, just store the URL as content (can be replaced with base64 fetch)
					contents = append(contents, MessageContent{Type: "image", Content: attachment.URL})
					break
				}
			}
		}
	}

	return NewMessage(m.Author.Username, contents)
}
