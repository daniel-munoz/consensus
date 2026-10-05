# consensus

![Version](https://img.shields.io/badge/version-1.2.0-blue)
![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8)
![License](https://img.shields.io/badge/license-BSD%203--Clause-green)

A CLI tool that sends prompts to multiple AI providers (OpenAI, Anthropic, Gemini) concurrently and saves their responses for comparison. Includes optional email notifications for real-time updates and a web UI for interactive use.

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

providers:
  - name: openai
    type: openai
    api_key_variable: OPENAI_API_KEY
    model: gpt-4o
  - name: gemini
    type: gemini
    api_key_variable: GEMINI_API_KEY
    model: gemini-2.0-flash
  - name: anthropic
    type: anthropic
    api_key_variable: ANTHROPIC_API_KEY
    model: claude-4-sonnet-20250514
    max_tokens: 64000

prompt_provider: openai
response_providers:
  - openai
  - gemini
  - anthropic
```

### Configuration Options

**Providers**: Configure which AI providers are available and their settings
- `name`: Unique identifier for the provider
- `type`: Provider type (`openai`, `gemini`, `anthropic`)
- `api_key_variable`: Environment variable containing the API key
- `model`: Model to use for this provider
- `max_tokens`: Optional token limit (primarily for Anthropic)
- `base_url`: Optional custom API endpoint

**Provider Selection**:
- `prompt_provider`: Which provider to use for optimizing prompts (default: `openai`)
- `response_providers`: List of providers to generate responses (default: all three)

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
# Email responses (full flag names)
go run main.go -prompt "Compare the pros and cons of React vs Vue" --email-to "user1@example.com,user2@example.com"

# Email shorthand
go run main.go -p "What are the latest trends in AI?" -e "user1@example.com,user2@example.com"

# Save responses to the local responses/ directory instead of (or in addition to) email
go run main.go -p "Your prompt here" --save-files
go run main.go -p "Your prompt here" -sf -e "user@example.com"

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

Local response files are **not** written by default. In command line mode at least one output is required: pass `--email-to`/`-e`, `--save-files`/`-sf`, or both. Otherwise the tool exits with an error before calling any provider.

### Server Mode (Web UI)
```bash
# Start the server with web UI (default port 8080)
go run main.go --serve

# Use a custom port
go run main.go --serve --port 3000

# Keep server running for multiple requests (default shuts down after first request completes)
go run main.go --serve --multi-session

# Also save responses to the local responses/ directory
go run main.go --serve --save-files
```

The server mode automatically opens your default browser to the web interface. The UI allows you to:
- Enter prompts interactively
- Select which providers to use for responses
- Toggle master prompt optimization
- View responses from all providers side by side

## How it works

The `consensus` tool follows these steps:

1. Accepts user input via command line flags (`-prompt` or `-p`) or interactive stdin prompt
2. Uses configurable provider (default: OpenAI) with a master prompt (C.R.A.F.T. methodology) to optimize the user's request (can be disabled with `--no-master-prompt` or overridden with `--master-prompt-provider`)
3. Sends the optimized prompt concurrently to configured AI providers (configurable via `--response-providers` or config file)
4. When `--save-files` is set, saves all outputs to the `responses/` directory (created if missing) with UUID-based filenames:
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
- **Embedded Web UI**: Self-contained HTTP server with an embedded web interface for interactive use

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
go run main.go -prompt "Your prompt here" --save-files

# Start web UI server
go run main.go --serve

# Run tests
go test ./...

# Build binary
go build
```
