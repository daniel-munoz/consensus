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
	BaseURL        *string
}

// NewGemini creates a new Gemini instance with the given configuration.
func NewGemini(configName, apiKeyVariable, model string, baseURL *string) *Gemini {
	return &Gemini{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		BaseURL:        baseURL,
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
	var (
		clientConfig  *genai.ClientConfig
		contentConfig *genai.GenerateContentConfig
	)
	apiKey := os.Getenv(p.APIKeyVariable)
	if apiKey == "" {
		return "", fmt.Errorf("%s environment variable not set", p.APIKeyVariable)
	}

	ctx := context.Background()

	clientConfig = &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	}
	if p.BaseURL != nil {
		clientConfig.HTTPOptions = genai.HTTPOptions{BaseURL: *p.BaseURL}
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}

	if system != nil {
		contentConfig = &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: *system}},
			},
		}
	}

	result, err := client.Models.GenerateContent(
		context.Background(),
		p.Model,
		genai.Text(prompt),
		contentConfig,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	return result.Text(), nil
}
