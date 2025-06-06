# consensus

A simple CLI tool to send prompts to AI providers (OpenAI, Anthropic, Gemini).

## Requirements

- Go 1.24 or higher.
- Environment variables:
  - `OPENAI_API_KEY` for OpenAI API.
  - `ANTHROPIC_API_KEY` for Anthropic API.
  - `GEMINI_API_KEY` for Google Gemini API.

## Usage

Run the tool:

```bash
go run main.go
```

## How it works

The `consensus` tool follows these steps:

1. Read a prompt from standard input.
2. Use a `master prompt` to craft an optimized, detailed prompt via OpenAI.
3. Send the optimized prompt concurrently to multiple AI providers (OpenAI, Anthropic, and Gemini).
4. Save each provider’s response to a separate text file (e.g., `openai.txt`, `anthropic.txt`, `gemini.txt`) for easy comparison.

