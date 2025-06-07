# consensus

A CLI tool that sends prompts to multiple AI providers (OpenAI, Anthropic, Gemini) concurrently and saves their responses for comparison.

## Requirements

- Go 1.24 or higher
- Environment variables:
  - `OPENAI_API_KEY` for OpenAI API
  - `ANTHROPIC_API_KEY` for Anthropic API
  - `GEMINI_API_KEY` for Google Gemini API

## Usage

Run the tool:

```bash
go run main.go
```

The tool will prompt you to enter your request, then process it through all AI providers.

## How it works

The `consensus` tool follows these steps:

1. Prompts user for input via stdin
2. Uses OpenAI with a master prompt (C.R.A.F.T. methodology) to optimize the user's request
3. Sends the optimized prompt concurrently to all three AI providers
4. Saves all outputs to the `responses/` directory with UUID-based filenames:
   - `id-{uuid}-request.txt` - Original user request
   - `id-{uuid}-prompt.txt` - Optimized prompt created by OpenAI
   - `id-{uuid}-OpenAI.txt` - OpenAI's response
   - `id-{uuid}-Anthropic.txt` - Anthropic's response  
   - `id-{uuid}-Gemini.txt` - Gemini's response

## Architecture

- **Provider Interface**: All AI providers implement a consistent `Provider` interface
- **Output Manager**: Flexible output system supporting multiple writers (currently file-based)
- **UUID Sessions**: Each run generates a unique session ID for organized file storage
- **Concurrent Processing**: All AI providers are queried simultaneously for faster results

## Development

- **Build and run**: `go run main.go`
- **Run tests**: `go test`
- **Build binary**: `go build`