package ai

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// Gemini represents a Gemini AI client configuration.
type Gemini struct {
	ConfigName     string
	APIKeyVariable string
	Model          string
}

// NewGemini creates a new Gemini instance with the given configuration.
func NewGemini(configName, apiKeyVariable, model string) *Gemini {
	return &Gemini{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
	}
}

// Name returns the name of the Gemini configuration.
func (p Gemini) Name() string {
	return p.ConfigName
}

// Type returns the type of the AI client.
func (_ Gemini) Type() string {
	return "gemini"
}

// Send sends a prompt to the Gemini API and returns the generated response.
func (p Gemini) Send(prompt string, system *string) (string, error) {
	var config *genai.GenerateContentConfig
	apiKey := os.Getenv(p.APIKeyVariable)
	if apiKey == "" {
		return "", fmt.Errorf("%s environment variable not set", p.APIKeyVariable)
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
		p.Model,
		genai.Text(prompt),
		config,
	)

	return result.Text(), err
}
