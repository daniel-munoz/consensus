# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

- **Build and run**: `go run main.go`
- **Build and run with prompt**: `go run main.go -prompt "Your prompt here"` or `go run main.go -p "Your prompt here"`
- **Run tests**: `go test`
- **Build binary**: `go build`

## Project Architecture

This is a CLI tool that sends prompts to multiple AI providers (OpenAI, Anthropic, Gemini) concurrently and saves their responses to separate files for comparison.

### Core Flow
1. Reads user prompt from command line flags (`-prompt` or `-p`) or interactive stdin input
2. Uses OpenAI with a "master prompt" (defined in `const.go`) to optimize the user's prompt using the C.R.A.F.T. methodology
3. Sends the optimized prompt concurrently to all three AI providers
4. Saves each response to provider-named files (`openai.txt`, `anthropic.txt`, `gemini.txt`)
5. Optionally sends email notifications for each step (user request, optimized prompt, and AI responses)

### Key Components
- **main.go**: Entry point with concurrent orchestration logic
- **config.go**: Configuration loader that creates/manages YAML config file
- **ai/**: Provider implementations that conform to the `Provider` interface
- **output/**: Output writers for file and email delivery
  - **file.go**: File creation utility for saving responses
  - **email.go**: Email delivery for real-time notifications
- **const.go**: Contains the master prompt template using C.R.A.F.T. methodology

### Provider Interface
All AI providers implement:
```go
type Provider interface {
    Name() string
    Send(string, *string) (string, error)
}
```


### Configuration

The app uses a YAML configuration file for email settings. On first run, a default config file is created at:
- `$XDG_CONFIG_HOME/consensus/config.yml` (if XDG_CONFIG_HOME is set)
- `~/.config/consensus/config.yml` (on most systems)
- `config.yml` (fallback in current directory)

Default config.yml:
```yaml
email:
  smtp_host: smtp.gmail.com
  smtp_port: 587
  from_email: consensus.ai.25@gmail.com
  from_name: Consensus AI
  password_env_var: CONSENSUS_EMAIL_PASSWORD
  subject_prefix: "[Consensus AI]"
```

#### Environment Variables
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY` 
- `GEMINI_API_KEY`
- Email password (configurable via `password_env_var` in config, defaults to `CONSENSUS_EMAIL_PASSWORD`)

### Command Line Usage
- **Interactive mode**: `go run main.go` (prompts for input)
- **Direct prompt**: `go run main.go -prompt "Your prompt here"`
- **Shorthand**: `go run main.go -p "Your prompt here"`
- **With email notifications**: `go run main.go -prompt "Your prompt here" --email-to "user1@example.com,user2@example.com"`
- **Email shorthand**: `go run main.go -p "Your prompt here" -e "user1@example.com,user2@example.com"`

### Testing
Tests focus on input validation and error handling for missing API keys. Use `go test` to run the test suite.
