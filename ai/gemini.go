package ai

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type Gemini struct{}

func (_ Gemini) Name() string {
	return "gemini"
}

func (_ Gemini) Send(prompt string, system *string) (string, error) {
	var config *genai.GenerateContentConfig
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}

	if system != nil {
		config = &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: *system}},
			},
		}
	}

	result, err := client.Models.GenerateContent(
		context.Background(),
		"gemini-2.0-flash",
		genai.Text(prompt),
		config,
	)

	return result.Text(), err
}
