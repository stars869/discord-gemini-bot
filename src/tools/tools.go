package tools

import (
	"context"
	"fmt"
)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ReturnDisplay string `json:"return_display"`
}

// Tool is the base interface for all tools
type Tool interface {
	// Name returns the name of the tool
	Name() string
	
	// Description returns the description of the tool
	Description() string
	
	// Run executes the tool synchronously
	Run(ctx context.Context, args ...interface{}) (*ToolResult, error)
	
	// ARun executes the tool asynchronously
	ARun(ctx context.Context, args ...interface{}) (*ToolResult, error)
}

// BaseTool provides a base implementation for tools
type BaseTool struct {
	name        string
	description string
}

// NewBaseTool creates a new base tool
func NewBaseTool(name, description string) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
	}
}

// Name returns the name of the tool
func (bt *BaseTool) Name() string {
	return bt.name
}

// Description returns the description of the tool
func (bt *BaseTool) Description() string {
	return bt.description
}

// Run provides a default implementation that returns not implemented
func (bt *BaseTool) Run(ctx context.Context, args ...interface{}) (*ToolResult, error) {
	return nil, fmt.Errorf("tool %s does not support sync execution", bt.name)
}

// ARun provides a default implementation that returns not implemented
func (bt *BaseTool) ARun(ctx context.Context, args ...interface{}) (*ToolResult, error) {
	return nil, fmt.Errorf("tool %s does not support async execution", bt.name)
}
