# Discord Gemini Bot (Go Version)

A Discord bot powered by Google's Gemini AI, rewritten in Go for better performance and concurrency.

## Features

- **AI-Powered Responses**: Uses Google's Gemini 2.0 Flash model for intelligent conversations
- **Tool Integration**: Supports Google Search and URL fetching tools
- **Multimodal Support**: Can process images along with text
- **Conversation Memory**: Maintains context across conversations per channel
- **Discord Integration**: Responds when mentioned in Discord channels
- **Concurrent Processing**: Built with Go's excellent concurrency support

## Prerequisites

- Go 1.21 or later
- Discord Bot Token
- Google Gemini API Key
- Google Search API credentials (optional, for search functionality)

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd discord-gemini-bot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your actual API keys
   ```

4. Build the application:
   ```bash
   go build -o discord-gemini-bot src/main.go
   ```

## Configuration

Create a `.env` file in the root directory with the following variables:

```env
DISCORD_BOT_TOKEN=your_discord_bot_token_here
GEMINI_API_KEY=your_gemini_api_key_here
GOOGLE_API_KEY=your_google_api_key_here  # Optional
GOOGLE_CSE_ID=your_google_cse_id_here    # Optional
```

## Usage

### Running the Bot

```bash
go run src/main.go
```

Or using the built binary:

```bash
./discord-gemini-bot
```

### Example Usage

See `examples/example_usage.go` for a simple example of how to use the Gemini model directly:

```bash
go run examples/example_usage.go
```

## Project Structure

```
├── src/
│   ├── main.go              # Main application entry point
│   ├── agent/
│   │   └── agent.go         # Agent logic and tool coordination
│   ├── models/
│   │   ├── llm_model.go     # LLM interface definition
│   │   └── gemini.go        # Gemini model implementation
│   ├── tools/
│   │   ├── tools.go         # Tool interface and base implementation
│   │   ├── google_search.go # Google Search tool
│   │   └── url_fetch.go     # URL fetching tool
│   ├── types/
│   │   ├── message.go       # Message type definition
│   │   └── memory.go        # Conversation memory management
│   ├── utils/
│   │   └── utils.go         # Utility functions
│   └── prompts/
│       └── prompts.go       # System prompts and templates
├── examples/
│   └── example_usage.go     # Example usage of the models
├── go.mod                   # Go module file
├── .env.example             # Environment variables template
└── README.md               # This file
```

## How It Works

1. **Discord Integration**: The bot listens for mentions in Discord channels
2. **Message Processing**: When mentioned, it processes the message and any attached images
3. **AI Response**: Uses Google's Gemini model to generate intelligent responses
4. **Tool Usage**: Can use tools like Google Search and URL fetching when needed
5. **Memory Management**: Maintains conversation context per channel
6. **Response Delivery**: Handles Discord's message length limits by splitting long responses

## Tools

The bot supports the following tools:

- **Google Search**: Searches Google for information
- **URL Fetch**: Fetches content from web URLs

## Development

### Adding New Tools

1. Create a new file in `src/tools/`
2. Implement the `Tool` interface
3. Add the tool to the `toolList` in `main.go`

### Adding New Models

1. Create a new file in `src/models/`
2. Implement the `LLMModel` interface
3. Update the initialization in `main.go`

## Performance Benefits of Go Version

- **Concurrent Processing**: Better handling of multiple Discord channels simultaneously
- **Memory Efficiency**: Lower memory usage compared to Python version
- **Faster Startup**: Compiled binary starts faster than Python script
- **Better Resource Management**: Explicit resource cleanup and management

## License

This project is licensed under the MIT License.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Troubleshooting

### Common Issues

1. **"Failed to create genai client"**: Check your `GEMINI_API_KEY` is correct
2. **"Discord connection failed"**: Verify your `DISCORD_BOT_TOKEN` is valid
3. **"Google Search not working"**: Ensure `GOOGLE_API_KEY` and `GOOGLE_CSE_ID` are set

### Debug Mode

To enable debug logging, set the log level:

```go
log.SetLevel(log.DebugLevel)
```

### Memory Usage

The bot uses a sliding window for conversation memory (default: 20 messages per channel). Adjust `MEMORY_WINDOW_SIZE` in `main.go` if needed.
