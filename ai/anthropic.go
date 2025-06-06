package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Anthropic struct{}

func (_ Anthropic) Name() string {
	return "anthropic"
}

func (_ Anthropic) Send(prompt string, system *string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	params := anthropic.MessageNewParams{

		Model:     anthropic.ModelClaude3_5Sonnet20241022,
		MaxTokens: int64(8192),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	if system != nil {
		params.System = []anthropic.TextBlockParam{{Text: *system}}
	}

	message, err := client.Messages.New(context.Background(), params)

	if err != nil {
		return "", fmt.Errorf("failed to send message to Anthropic: %w", err)
	}

	return message.Content[0].AsText().Text, nil
}
