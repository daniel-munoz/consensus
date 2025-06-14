# consensus

A CLI tool that sends prompts to multiple AI providers (OpenAI, Anthropic, Gemini) concurrently and saves their responses for comparison. Includes optional email notifications for real-time updates.

## Requirements

- Go 1.24 or higher
- Environment variables:
  - `OPENAI_API_KEY` for OpenAI API
  - `ANTHROPIC_API_KEY` for Anthropic API
  - `GEMINI_API_KEY` for Google Gemini API

## Configuration

The app uses a YAML configuration file for email settings. On first run, a default config file is created at:
- `$XDG_CONFIG_HOME/consensus/config.yml` (if XDG_CONFIG_HOME is set)
- `~/.config/consensus/config.yml` (on most systems)  
- `config.yml` (fallback in current directory)

### Default Configuration
```yaml
email:
  smtp_host: smtp.gmail.com
  smtp_port: 587
  from_email: consensus.ai.25@gmail.com
  from_name: Consensus AI
  password_env_var: CONSENSUS_EMAIL_PASSWORD
  subject_prefix: "[Consensus AI]"
```

### Optional Email Notifications
- Environment variable for email password (configurable via `password_env_var` in config, defaults to `CONSENSUS_EMAIL_PASSWORD`)

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

# Disable master prompt optimization
go run main.go -p "Your prompt here" --no-master-prompt
go run main.go -p "Your prompt here" -nmp

# Override which provider handles master prompt optimization
go run main.go -p "Your prompt here" --master-prompt-provider "anthropic"
go run main.go -p "Your prompt here" -mpp "gemini"

# Override which providers generate responses
go run main.go -p "Your prompt here" --response-providers "openai,gemini"
go run main.go -p "Your prompt here" -rp "anthropic,openai"

# Combined configuration overrides
go run main.go -p "Your prompt here" -mpp "gemini" -rp "openai,anthropic" -e "user@example.com"
```

## How it works

The `consensus` tool follows these steps:

1. Accepts user input via command line flags (`-prompt` or `-p`) or interactive stdin prompt
2. Uses configurable provider (default: OpenAI) with a master prompt (C.R.A.F.T. methodology) to optimize the user's request (can be disabled with `--no-master-prompt` or overridden with `--master-prompt-provider`)
3. Sends the optimized prompt concurrently to configured AI providers (configurable via `--response-providers` or config file)
4. Saves all outputs to the `responses/` directory with UUID-based filenames:
   - `id-{uuid}-request.txt` - Original user request
   - `id-{uuid}-prompt.txt` - Optimized prompt created by master prompt provider
   - `id-{uuid}-{Provider}.txt` - Each provider's response (e.g., OpenAI.txt, Anthropic.txt, Gemini.txt)
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
