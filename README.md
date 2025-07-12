# Discord Gemini Bot (Go Version)

A high-performance Discord bot powered by Google's Gemini AI, written in Go for superior concurrency and resource efficiency.

## 🚀 Features

- **AI-Powered Responses**: Uses Google's Gemini 2.0 Flash model for intelligent conversations
- **Tool Integration**: Supports Google Search and URL fetching tools
- **Multimodal Support**: Can process images along with text
- **Conversation Memory**: Maintains context across conversations per channel
- **Discord Integration**: Responds when mentioned in Discord channels
- **High Performance**: Built with Go's excellent concurrency support
- **Resource Efficient**: Lower memory usage and faster startup compared to Python

## 📋 Prerequisites

- Go 1.23 or later
- Discord Bot Token
- Google Gemini API Key
- Google Search API credentials (optional, for search functionality)

## 🛠 Installation

### 1. Clone the repository
```bash
git clone <repository-url>
cd discord-gemini-bot
```

### 2. Install dependencies
```bash
make deps
# or manually:
go mod download
go mod tidy
```

### 3. Set up environment variables
```bash
cp .env.template .env
# Edit .env with your actual API keys
```

### 4. Build the application
```bash
make build
# or manually:
go build -o discord-gemini-bot src/main.go
```

## ⚙️ Configuration

Create a `.env` file in the root directory with the following variables:

```env
# Required
DISCORD_BOT_TOKEN=your_discord_bot_token_here
GEMINI_API_KEY=your_gemini_api_key_here

# Optional (for enhanced functionality)
GOOGLE_API_KEY=your_google_api_key_here
GOOGLE_CSE_ID=your_google_cse_id_here
```

### Getting API Keys

1. **Discord Bot Token**: 
   - Go to [Discord Developer Portal](https://discord.com/developers/applications)
   - Create a new application and bot
   - Copy the bot token

2. **Gemini API Key**: 
   - Visit [Google AI Studio](https://aistudio.google.com/app/apikey)
   - Create a new API key

3. **Google Search API** (Optional):
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Enable Custom Search JSON API
   - Create credentials and a Custom Search Engine

## 🚀 Usage

### Running the Bot

```bash
# Using Make
make run

# Or directly
./discord-gemini-bot

# Or with go run
go run src/main.go
```

### Testing the Setup

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests directly with go
go test -v ./tests/

# Run example usage
make example
```

### Development Commands

```bash
# Build only
make build

# Clean build artifacts
make clean

# Check code quality
make check

# View all available commands
make help
```

## 📁 Project Structure

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
├── tests/
│   └── models_test.go       # Comprehensive model tests
├── examples/
│   └── example_usage.go     # Example usage of the models
├── Makefile                 # Build and run commands
├── go.mod                   # Go module file
├── .env.template            # Environment variables template
└── README.md               # This file
```

## 🔧 How It Works

1. **Discord Integration**: The bot listens for mentions in Discord channels
2. **Message Processing**: When mentioned, it processes the message and any attached images
3. **AI Response**: Uses Google's Gemini model to generate intelligent responses
4. **Tool Usage**: Can use tools like Google Search and URL fetching when needed
5. **Memory Management**: Maintains conversation context per channel
6. **Response Delivery**: Handles Discord's message length limits by splitting long responses

## 🛠 Available Tools

The bot supports the following tools:

- **Google Search**: Searches Google for information
- **URL Fetch**: Fetches content from web URLs

## 🔨 Development

### Adding New Tools

1. Create a new file in `src/tools/`
2. Implement the `Tool` interface:
   ```go
   type Tool interface {
       Name() string
       Description() string
       ARun(ctx context.Context, args ...interface{}) (*ToolResult, error)
   }
   ```
3. Add the tool to the `toolList` in `main.go`

### Adding New Models

1. Create a new file in `src/models/`
2. Implement the `LLMModel` interface:
   ```go
   type LLMModel interface {
       GenerateAsync(ctx context.Context, prompt string, images []map[string]interface{}) (string, error)
       GenerateWithHistoryAsync(ctx context.Context, messages []*types.Message) (string, error)
       SetSystemPrompt(systemPrompt string)
   }
   ```
3. Update the initialization in `main.go`

## 🚀 Performance Benefits of Go Version

- **Concurrent Processing**: Excellent handling of multiple Discord channels simultaneously
- **Memory Efficiency**: Significantly lower memory usage compared to Python version
- **Faster Startup**: Compiled binary starts much faster than Python script
- **Better Resource Management**: Explicit resource cleanup and management
- **Type Safety**: Compile-time error checking prevents runtime issues
- **Single Binary**: Easy deployment with no dependency management

## 📊 Configuration Options

You can modify these constants in `src/main.go`:

```go
const (
    MEMORY_WINDOW_SIZE         = 20   // Number of messages to remember per channel
    DISCORD_MAX_MESSAGE_LENGTH = 2000 // Discord's message length limit
)
```

## 🐛 Troubleshooting

### Common Issues

1. **"Failed to create genai client"**: 
   - Check your `GEMINI_API_KEY` is correct
   - Ensure you have proper internet connectivity

2. **"Discord connection failed"**: 
   - Verify your `DISCORD_BOT_TOKEN` is valid
   - Make sure the bot has proper permissions in your Discord server

3. **"Google Search not working"**: 
   - Ensure `GOOGLE_API_KEY` and `GOOGLE_CSE_ID` are set correctly
   - Check that the Custom Search JSON API is enabled

4. **"Permission denied"**: 
   - Make sure the binary has execute permissions: `chmod +x discord-gemini-bot`

### Debug Mode

To enable verbose logging, you can modify the log level in the code or set environment variables.

### Memory Usage

The bot uses a sliding window for conversation memory (default: 20 messages per channel). Adjust `MEMORY_WINDOW_SIZE` in `main.go` if needed.

## 📝 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📞 Support

If you encounter any issues or have questions:

1. Check the [Troubleshooting](#-troubleshooting) section
2. Look through existing GitHub issues
3. Create a new issue with detailed information about your problem

## 🙏 Acknowledgments

- Google for the Gemini API
- Discord for their excellent bot API
- The Go community for amazing libraries and tools
