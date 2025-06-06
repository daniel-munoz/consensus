package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OpenAI struct{}

func (_ OpenAI) Name() string {
	return "openai"
}

func (_ OpenAI) Send(prompt string, system *string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	messages := []map[string]string{}
	if system != nil {
		messages = append(messages, map[string]string{"role": "developer", "content": *system})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})
	reqBody := map[string]any{"model": "gpt-4o", "messages": messages}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
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
