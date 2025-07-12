package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// URLFetchTool implements URL content fetching functionality
type URLFetchTool struct {
	*BaseTool
	client *http.Client
}

// NewURLFetchTool creates a new URL fetch tool
func NewURLFetchTool() *URLFetchTool {
	return &URLFetchTool{
		BaseTool: NewBaseTool(
			"url_fetch",
			"Fetches the content of a given URL. Input should be a valid URL string.",
		),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ARun executes the URL fetch tool asynchronously
func (uft *URLFetchTool) ARun(ctx context.Context, args ...interface{}) (*ToolResult, error) {
	if len(args) == 0 {
		return &ToolResult{ReturnDisplay: "Error: No URL provided"}, nil
	}

	urlStr, ok := args[0].(string)
	if !ok {
		return &ToolResult{ReturnDisplay: "Error: URL must be a string"}, nil
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error creating request for URL %s: %v", urlStr, err)}, nil
	}

	// Set a reasonable user agent
	req.Header.Set("User-Agent", "Discord-Gemini-Bot/1.0")

	// Make the request
	resp, err := uft.client.Do(req)
	if err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error fetching URL %s: %v", urlStr, err)}, nil
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error fetching URL %s: HTTP %d", urlStr, resp.StatusCode)}, nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error reading response from URL %s: %v", urlStr, err)}, nil
	}

	// Limit the content to first 1000 characters to avoid overwhelming the LLM
	content := string(body)
	if len(content) > 1000 {
		content = content[:1000] + "..."
	}

	return &ToolResult{ReturnDisplay: fmt.Sprintf("Content from %s:\n%s", urlStr, content)}, nil
}
