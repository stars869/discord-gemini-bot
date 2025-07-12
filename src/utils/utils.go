package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SupportedImageTypes contains the supported image MIME types
var SupportedImageTypes = []string{
	"image/png",
	"image/jpeg", 
	"image/webp",
	"image/gif",
}

// ImageData represents image data with MIME type
type ImageData struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GetImageAsBase64 fetches an image from a URL and returns its base64 encoded string and MIME type
func GetImageAsBase64(ctx context.Context, url string) (*ImageData, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("no content type specified")
	}

	// Check if the content type is supported
	supported := false
	for _, supportedType := range SupportedImageTypes {
		if contentType == supportedType {
			supported = true
			break
		}
	}

	if !supported {
		return nil, fmt.Errorf("unsupported image type: %s", contentType)
	}

	// Read the image data
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading image data: %w", err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)

	return &ImageData{
		MIMEType: contentType,
		Data:     base64Data,
	}, nil
}

// SplitLongText splits a long text into chunks, prioritizing newline characters
func SplitLongText(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	var chunks []string
	var currentChunk strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		// Check if adding the current line would exceed the max length
		newlineChar := 0
		if currentChunk.Len() > 0 {
			newlineChar = 1 // Account for the newline character
		}

		if currentChunk.Len()+len(line)+newlineChar <= maxLength {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n")
			}
			currentChunk.WriteString(line)
		} else {
			// If the current chunk is not empty, add it to chunks
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}

			// Handle the line that didn't fit
			if len(line) > maxLength {
				// Split this long line into sub-chunks
				for i := 0; i < len(line); i += maxLength {
					end := i + maxLength
					if end > len(line) {
						end = len(line)
					}
					chunks = append(chunks, line[i:end])
				}
			} else {
				// Start a new chunk with this line
				currentChunk.WriteString(line)
			}
		}
	}

	// Add any remaining content in currentChunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
