package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// GoogleSearchTool implements Google search functionality
type GoogleSearchTool struct {
	*BaseTool
	apiKey string
	cseID  string
	url    string
}

// GoogleSearchResult represents a single search result
type GoogleSearchResult struct {
	Items []struct {
		Title   string `json:"title"`
		Snippet string `json:"snippet"`
		Link    string `json:"link"`
	} `json:"items"`
}

// NewGoogleSearchTool creates a new Google search tool
func NewGoogleSearchTool() *GoogleSearchTool {
	return &GoogleSearchTool{
		BaseTool: NewBaseTool(
			"google_search",
			"Searches Google for the given query.",
		),
		apiKey: os.Getenv("GOOGLE_API_KEY"),
		cseID:  os.Getenv("GOOGLE_CSE_ID"),
		url:    "https://www.googleapis.com/customsearch/v1",
	}
}

// ARun executes the Google search tool asynchronously
func (gst *GoogleSearchTool) ARun(ctx context.Context, args ...interface{}) (*ToolResult, error) {
	if len(args) == 0 {
		return &ToolResult{ReturnDisplay: "Error: No query provided"}, nil
	}

	query, ok := args[0].(string)
	if !ok {
		return &ToolResult{ReturnDisplay: "Error: Query must be a string"}, nil
	}

	if gst.apiKey == "" || gst.cseID == "" {
		return &ToolResult{ReturnDisplay: "Error: Google API key or CSE ID not configured"}, nil
	}

	// Build the request URL
	params := url.Values{}
	params.Set("key", gst.apiKey)
	params.Set("cx", gst.cseID)
	params.Set("q", query)

	requestURL := fmt.Sprintf("%s?%s", gst.url, params.Encode())

	// Make the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error creating request: %v", err)}, nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error making request: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error: HTTP %d", resp.StatusCode)}, nil
	}

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error reading response: %v", err)}, nil
	}

	var result GoogleSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return &ToolResult{ReturnDisplay: fmt.Sprintf("Error parsing response: %v", err)}, nil
	}

	// Extract snippets
	var snippets []string
	for _, item := range result.Items {
		snippets = append(snippets, item.Snippet)
	}

	if len(snippets) == 0 {
		return &ToolResult{ReturnDisplay: "No results found"}, nil
	}

	return &ToolResult{ReturnDisplay: strings.Join(snippets, "\n")}, nil
}
