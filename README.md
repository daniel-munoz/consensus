# consensus

A simple CLI tool to send prompts to AI providers (OpenAI, Anthropic, Gemini).

## Requirements

- Go 1.24 or higher.
- Environment variables:
  - `OPENAI_API_KEY` for OpenAI API.
  - `ANTHROPIC_API_KEY` for Anthropic API.
  - `GOOGLE_API_KEY` for Google Gemini API.

## Usage

Run the tool:

```bash
go run main.go
```

You'll be prompted to select a provider, enter a model name, and input your prompt. The response from the selected AI model will then be displayed.

> **Note:** When using Anthropic, your prompt will be automatically wrapped with the required prefixes (`\n\nHuman:` and `\n\nAssistant:`) as per the Anthropic API specification.