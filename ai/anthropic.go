package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Anthropic represents a client for interacting with the Anthropic AI service.
type Anthropic struct {
	ConfigName     string
	APIKeyVariable string
	Model          string
	MaxTokens      int64
	BaseURL        *string
}

// NewAnthropic creates a new instance of the Anthropic AI client with the specified configuration.
func NewAnthropic(configName, apiKeyVariable, model string, maxTokens int64, baseURL *string) Anthropic {
	return Anthropic{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		MaxTokens:      maxTokens,
		BaseURL:        baseURL,
	}
}

// Name returns the name of the AI client configuration.
func (p Anthropic) Name() string {
	return p.ConfigName
}

// Type returns the type of the AI provider.
func (_ Anthropic) Type() string {
	return "anthropic"
}

// Send sends a prompt to the Anthropic AI model and returns the response.
func (p Anthropic) Send(prompt string, system *string) (string, error) {
	apiKey := os.Getenv(p.APIKeyVariable)
	if apiKey == "" {
		return "", fmt.Errorf("%s environment variable not set", p.APIKeyVariable)
	}

	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if p.BaseURL != nil {
		options = append(options, option.WithBaseURL(*p.BaseURL))
	}

	client := anthropic.NewClient(
		options...,
	)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.Model),
		MaxTokens: p.MaxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	if system != nil {
		params.System = []anthropic.TextBlockParam{{Text: *system}}
	}

	stream := client.Messages.NewStreaming(context.Background(), params)

	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		err := message.Accumulate(event)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve message from Anthropic: %w", err)
		}
	}

	if stream.Err() != nil {
		return "", fmt.Errorf("failed to retrieve message from Anthropic: %w", stream.Err())
	}

	return message.Content[0].AsText().Text, nil
}
