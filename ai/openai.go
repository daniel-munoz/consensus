package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	messages := []map[string]string{}
	if system != nil {
		messages = append(messages, map[string]string{"role": "developer", "content": *system})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})
	reqBody := map[string]any{"model": p.Model, "messages": messages}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	complationsURL := "https://api.openai.com/v1/chat/completions"
	if p.BaseURL != nil {
		complationsURL = *p.BaseURL + "/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", complationsURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}
	var respBody struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}
	if len(respBody.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}
	return respBody.Choices[0].Message.Content, nil
}
