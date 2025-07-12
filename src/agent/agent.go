package agent

import (
	"context"
	"discord-gemini-bot/src/models"
	"discord-gemini-bot/src/prompts"
	"discord-gemini-bot/src/tools"
	"discord-gemini-bot/src/types"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// Agent represents the main agent class that handles user interactions
type Agent struct {
	model  models.LLMModel
	memory *types.ConversationMemory
	tools  map[string]tools.Tool
}

// NewAgent creates a new agent instance
func NewAgent(model models.LLMModel, memory *types.ConversationMemory, toolList []tools.Tool) *Agent {
	// Create tools map
	toolsMap := make(map[string]tools.Tool)
	for _, tool := range toolList {
		toolsMap[tool.Name()] = tool
	}

	agent := &Agent{
		model:  model,
		memory: memory,
		tools:  toolsMap,
	}

	// Set up system prompt
	toolsString := agent.getToolsString()
	toolNames := agent.getToolNames()
	systemPrompt := fmt.Sprintf(prompts.GetAgentSystemPromptTemplate(), toolsString, toolNames)
	model.SetSystemPrompt(systemPrompt)

	log.Printf("System prompt set for agent")
	return agent
}

// getToolNames returns a comma-separated string of tool names
func (a *Agent) getToolNames() string {
	var names []string
	for name := range a.tools {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// getToolsString returns a formatted string of all available tools
func (a *Agent) getToolsString() string {
	var toolDescriptions []string
	for name, tool := range a.tools {
		toolDescriptions = append(toolDescriptions, fmt.Sprintf("%s: %s", name, tool.Description()))
	}
	return strings.Join(toolDescriptions, "\n")
}

// AddMessage adds a message to the agent's memory
func (a *Agent) AddMessage(message *types.Message) {
	a.memory.AddMessage(message)
}

// GetResponse gets a response from the agent
func (a *Agent) GetResponse(ctx context.Context) (string, error) {
	// Get conversation history
	messages := a.memory.GetHistory()

	// Generate response using the model
	response, err := a.model.GenerateWithHistoryAsync(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("error generating response: %w", err)
	}

	log.Printf("Model's raw response: %s", response)

	// Check for tool use
	toolRegex := regexp.MustCompile(`Action: (\w+)\nAction Input: (.*)`)
	matches := toolRegex.FindStringSubmatch(response)

	if len(matches) >= 3 {
		toolName := strings.TrimSpace(matches[1])
		toolInput := strings.TrimSpace(matches[2])

		log.Printf("Tool use detected: %s with input %s", toolName, toolInput)

		tool, exists := a.tools[toolName]
		if exists {
			// Execute the tool
			toolResult, err := tool.ARun(ctx, toolInput)
			if err != nil {
				log.Printf("Error executing tool %s: %v", toolName, err)
				observation := fmt.Sprintf("Tool %s failed: %v", toolName, err)
				aiMsg := types.NewMessage("AI", []types.MessageContent{{Type: "text", Content: observation}})
				a.AddMessage(aiMsg)
			} else {
				observation := fmt.Sprintf("Tool %s used. Observation: %s", toolName, toolResult.ReturnDisplay)
				log.Printf("Tool observation: %s", observation)
				aiMsg := types.NewMessage("AI", []types.MessageContent{{Type: "text", Content: observation}})
				a.AddMessage(aiMsg)

				// Get a new response with the tool's output using conversation history
				history := a.memory.GetHistory()
				response, err = a.model.GenerateWithHistoryAsync(ctx, history)
				if err != nil {
					return "", fmt.Errorf("error generating follow-up response: %w", err)
				}

				log.Printf("Model's raw response after tool use: %s", response)
			}
		} else {
			log.Printf("Tool %s not found", toolName)
		}
	}

	// Add AI response to memory
	aiMsg := types.NewMessage("AI", []types.MessageContent{{Type: "text", Content: response}})
	a.AddMessage(aiMsg)

	log.Printf("Agent's final response: %s", response)
	return response, nil
}
