# consensus

A CLI tool that sends prompts to multiple AI providers (OpenAI, Anthropic, Gemini) concurrently and saves their responses for comparison. Includes optional email notifications for real-time updates.

## Requirements

- Go 1.24 or higher
- Environment variables:
  - `OPENAI_API_KEY` for OpenAI API
  - `ANTHROPIC_API_KEY` for Anthropic API
  - `GEMINI_API_KEY` for Google Gemini API

### Optional Email Notifications
- `CONSENSUS_EMAIL_PASSWORD` - App password for consensus.ai.25@gmail.com

## Usage

### Interactive Mode
```bash
go run main.go
```
The tool will prompt you to enter your request, then process it through all AI providers.

### Command Line Mode
```bash
# Using full flag name
go run main.go -prompt "Compare the pros and cons of React vs Vue"

# Using shorthand
go run main.go -p "What are the latest trends in AI?"

# With email notifications
go run main.go -prompt "Your prompt here" --email-to "user1@example.com,user2@example.com"

# Email shorthand
go run main.go -p "Your prompt here" -e "user1@example.com,user2@example.com"
```

## How it works

The `consensus` tool follows these steps:

1. Accepts user input via command line flags (`-prompt` or `-p`) or interactive stdin prompt
2. Uses OpenAI with a master prompt (C.R.A.F.T. methodology) to optimize the user's request
3. Sends the optimized prompt concurrently to all three AI providers
4. Saves all outputs to the `responses/` directory with UUID-based filenames:
   - `id-{uuid}-request.txt` - Original user request
   - `id-{uuid}-prompt.txt` - Optimized prompt created by OpenAI
   - `id-{uuid}-OpenAI.txt` - OpenAI's response
   - `id-{uuid}-Anthropic.txt` - Anthropic's response  
   - `id-{uuid}-Gemini.txt` - Gemini's response
5. Optionally sends HTML email notifications for each step when email is configured

## Architecture

- **Provider Interface**: All AI providers implement a consistent `Provider` interface
- **Output Manager**: Flexible output system supporting multiple writers (file and email)
- **UUID Sessions**: Each run generates a unique session ID for organized file storage and email tracking
- **Concurrent Processing**: All AI providers are queried simultaneously for faster results
- **Email Notifications**: Optional real-time email updates with HTML formatting and provider-specific styling

## Development

### Using Make (Recommended)
```bash
# Build and run
make run

# Run with a specific prompt
make run PROMPT="Your prompt here"

# Run tests
make test

# Build binary
make build

# Clean build artifacts
make clean
```

### Direct Go Commands
```bash
# Build and run
go run main.go

# Build and run with prompt
go run main.go -prompt "Your prompt here"

# Run tests
go test ./...

# Build binary
go build
```