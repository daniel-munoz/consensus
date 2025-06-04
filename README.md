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

You'll be prompted to select a provider, enter a model name, and input your prompt. The response from the selected AI model will then be displayed.

> **Note:** Anthropic prompts are sent directly, without adding `\n\nHuman:` or `\n\nAssistant:` prefixes.