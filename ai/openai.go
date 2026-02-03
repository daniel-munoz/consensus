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
	clientFactory  OpenAIClientFactory // nil = use default
}

// realOpenAIClient wraps the actual OpenAI SDK client
type realOpenAIClient struct {
	apiKey  string
	baseURL *string
}

func (c *realOpenAIClient) CreateResponse(ctx context.Context, model, input string, instructions *string) (string, error) {
	var oaiClient openai.Client
	if c.baseURL != nil && *c.baseURL != "" {
		oaiClient = openai.NewClient(
			option.WithAPIKey(c.apiKey),
			option.WithBaseURL(*c.baseURL),
		)
	} else {
		oaiClient = openai.NewClient(
			option.WithAPIKey(c.apiKey),
		)
	}

	params := responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(input),
		},
	}

	if instructions != nil {
		params.Instructions = openai.String(*instructions)
	}

	resp, err := oaiClient.Responses.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create response: %w", err)
	}

	return resp.OutputText(), nil
}

func defaultOpenAIClientFactory(apiKey string, baseURL *string) OpenAIClient {
	return &realOpenAIClient{apiKey: apiKey, baseURL: baseURL}
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

// NewOpenAIWithClient creates an OpenAI instance with a custom client factory (for testing)
func NewOpenAIWithClient(configName, apiKeyVariable, model string, baseURL *string, factory OpenAIClientFactory) OpenAI {
	return OpenAI{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		BaseURL:        baseURL,
		clientFactory:  factory,
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

	factory := p.clientFactory
	if factory == nil {
		factory = defaultOpenAIClientFactory
	}

	client := factory(apiKey, p.BaseURL)
	return client.CreateResponse(context.Background(), p.Model, prompt, system)
}
