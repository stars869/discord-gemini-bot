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

// GetResponse gets a response from the agent
func (a *Agent) GetResponse(ctx context.Context, author, message string, images []map[string]interface{}) (string, error) {
	// Add user message to memory
	a.memory.AddMessage(author, message)
	
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
				a.memory.AddMessage("AI", observation)
			} else {
				observation := fmt.Sprintf("Tool %s used. Observation: %s", toolName, toolResult.ReturnDisplay)
				log.Printf("Tool observation: %s", observation)
				a.memory.AddMessage("AI", observation)
				
				// Get a new response with the tool's output
				history := a.memory.GetHistory()
				prompt := fmt.Sprintf("%s\n\nTOOLS:\n------\n%s\n\nPrevious conversation history:\n%s\n\nNew input: %s: %s\nFinal Answer:", 
					prompts.GetAgentSystemPromptTemplate(), 
					a.getToolsString(), 
					a.formatHistory(history), 
					author, 
					message)
				
				log.Printf("Prompt sent to model after tool use")
				
				// Generate follow-up response
				response, err = a.model.GenerateAsync(ctx, prompt, images)
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
	a.memory.AddMessage("AI", response)
	
	log.Printf("Agent's final response: %s", response)
	return response, nil
}

// formatHistory formats the conversation history for the prompt
func (a *Agent) formatHistory(messages []*types.Message) string {
	var formatted []string
	for _, msg := range messages {
		formatted = append(formatted, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
	}
	return strings.Join(formatted, "\n")
}
