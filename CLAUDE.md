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

### Key Components
- **main.go**: Entry point with concurrent orchestration logic
- **ai/**: Provider implementations that conform to the `Provider` interface
- **file/create.go**: File creation utility for saving responses
- **const.go**: Contains the master prompt template using C.R.A.F.T. methodology

### Provider Interface
All AI providers implement:
```go
type Provider interface {
    Name() string
    Send(string, *string) (string, error)
}
```

### Environment Variables Required
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY` 
- `GEMINI_API_KEY`

### Command Line Usage
- **Interactive mode**: `go run main.go` (prompts for input)
- **Direct prompt**: `go run main.go -prompt "Your prompt here"`
- **Shorthand**: `go run main.go -p "Your prompt here"`

### Testing
Tests focus on input validation and error handling for missing API keys. Use `go test` to run the test suite.