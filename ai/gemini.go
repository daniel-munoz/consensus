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
	clientFactory  GeminiClientFactory // nil = use default
}

// realGeminiClient wraps the actual Gemini SDK client
type realGeminiClient struct {
	apiKey  string
	baseURL *string
}

func (c *realGeminiClient) GenerateContent(ctx context.Context, model, prompt string, system *string) (string, error) {
	clientConfig := &genai.ClientConfig{
		APIKey:  c.apiKey,
		Backend: genai.BackendGeminiAPI,
	}
	if c.baseURL != nil {
		clientConfig.HTTPOptions = genai.HTTPOptions{BaseURL: *c.baseURL}
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}

	var contentConfig *genai.GenerateContentConfig
	if system != nil {
		contentConfig = &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: *system}},
			},
		}
	}

	result, err := client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		contentConfig,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	return result.Text(), nil
}

func defaultGeminiClientFactory(apiKey string, baseURL *string) GeminiClient {
	return &realGeminiClient{apiKey: apiKey, baseURL: baseURL}
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

// NewGeminiWithClient creates a Gemini instance with a custom client factory (for testing)
func NewGeminiWithClient(configName, apiKeyVariable, model string, baseURL *string, factory GeminiClientFactory) *Gemini {
	return &Gemini{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		BaseURL:        baseURL,
		clientFactory:  factory,
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
	apiKey := os.Getenv(p.APIKeyVariable)
	if apiKey == "" {
		return "", fmt.Errorf("%s environment variable not set", p.APIKeyVariable)
	}

	factory := p.clientFactory
	if factory == nil {
		factory = defaultGeminiClientFactory
	}

	client := factory(apiKey, p.BaseURL)
	return client.GenerateContent(context.Background(), p.Model, prompt, system)
}
