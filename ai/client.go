package ai

import "context"

// OpenAIClient abstracts the OpenAI SDK client for testing
type OpenAIClient interface {
	CreateResponse(ctx context.Context, model, input string, instructions *string) (string, error)
}

// OpenAIClientFactory creates an OpenAIClient instance
type OpenAIClientFactory func(apiKey string, baseURL *string) OpenAIClient

// AnthropicClient abstracts the Anthropic SDK client for testing
type AnthropicClient interface {
	CreateMessage(ctx context.Context, model string, maxTokens int64, prompt string, system *string) (string, error)
}

// AnthropicClientFactory creates an AnthropicClient instance
type AnthropicClientFactory func(apiKey string, baseURL *string) AnthropicClient

// GeminiClient abstracts the Gemini SDK client for testing
type GeminiClient interface {
	GenerateContent(ctx context.Context, model, prompt string, system *string) (string, error)
}

// GeminiClientFactory creates a GeminiClient instance
type GeminiClientFactory func(apiKey string, baseURL *string) GeminiClient
