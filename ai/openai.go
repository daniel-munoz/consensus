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
	MaxTokens      *int64 // nil = use the server default
	BaseURL        *string
	clientFactory  OpenAIClientFactory // nil = use default
}

// realOpenAIClient wraps the actual OpenAI SDK client
type realOpenAIClient struct {
	apiKey  string
	baseURL *string
}

func (c *realOpenAIClient) CreateResponse(ctx context.Context, model, input string, instructions *string, maxOutputTokens *int64) (string, error) {
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

	if maxOutputTokens != nil {
		params.MaxOutputTokens = openai.Int(*maxOutputTokens)
	}

	resp, err := oaiClient.Responses.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create response: %w", err)
	}

	return outputTextOrError(resp)
}

// outputTextOrError extracts the response text, failing when there is none and
// flagging responses that were cut short (e.g. by max_output_tokens).
func outputTextOrError(resp *responses.Response) (string, error) {
	text := resp.OutputText()
	incomplete := resp.Status == responses.ResponseStatusIncomplete

	if text == "" {
		if incomplete {
			return "", fmt.Errorf("response incomplete with no text (reason: %s)", resp.IncompleteDetails.Reason)
		}
		return "", fmt.Errorf("response contained no text (status: %s)", resp.Status)
	}

	if incomplete {
		text += fmt.Sprintf("\n\n[Response truncated: %s]", resp.IncompleteDetails.Reason)
	}

	return text, nil
}

func defaultOpenAIClientFactory(apiKey string, baseURL *string) OpenAIClient {
	return &realOpenAIClient{apiKey: apiKey, baseURL: baseURL}
}

// NewOpenAI creates a new OpenAI instance with the specified configuration.
func NewOpenAI(configName, apiKeyVariable, model string, maxTokens *int64, baseURL *string) OpenAI {
	return OpenAI{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		MaxTokens:      maxTokens,
		BaseURL:        baseURL,
	}
}

// NewOpenAIWithClient creates an OpenAI instance with a custom client factory (for testing)
func NewOpenAIWithClient(configName, apiKeyVariable, model string, maxTokens *int64, baseURL *string, factory OpenAIClientFactory) OpenAI {
	return OpenAI{
		ConfigName:     configName,
		APIKeyVariable: apiKeyVariable,
		Model:          model,
		MaxTokens:      maxTokens,
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
	return client.CreateResponse(context.Background(), p.Model, prompt, system, p.MaxTokens)
}
