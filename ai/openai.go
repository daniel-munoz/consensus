package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

// OpenAI represents an AI provider that interacts with the OpenAI API.
type OpenAI struct {
	ConfigName     string
	APIKeyVariable string
	Model          string
	BaseURL        *string
}

// NewOpenAI creates a new OpenAI instance with the specified configuration.
func NewOpenAI(configName, apiKeyVariable, model string, baseURL *string) OpenAI {
	return OpenAI{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		BaseURL:        baseURL,
	}
}

// Type returns the type of the AI provider.
func (_ OpenAI) Type() string {
	return "openai"
}

// Name returns the name of the AI provider.
func (p OpenAI) Name() string {
	return p.ConfigName
}

// Send sends a prompt to the OpenAI API and returns the response.
func (p OpenAI) Send(prompt string, system *string) (string, error) {
	apiKey := os.Getenv(p.APIKeyVariable)
	if apiKey == "" {
		return "", fmt.Errorf("%s environment variable not set", p.APIKeyVariable)
	}

	var oaiClient openai.Client
	if p.BaseURL != nil && *p.BaseURL != "" {
		oaiClient = openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(*p.BaseURL),
		)
	} else {
		oaiClient = openai.NewClient(
			option.WithAPIKey(apiKey),
		)
	}

	params := responses.ResponseNewParams{
		Model: p.Model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	}

	if system != nil {
		params.Instructions = openai.String(*system)
	}

	resp, err := oaiClient.Responses.New(context.Background(), params)
	if err != nil {
		return "", fmt.Errorf("failed to create response: %w", err)
	}

	return resp.OutputText(), nil
}
