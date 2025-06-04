package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"google.golang.org/genai"
)

func main() {
	provider := selectProvider()
	model := selectModel(provider)
	prompt := readPrompt()

	var (
		response string
		err      error
	)
	switch provider {
	case "OpenAI":
		response, err = sendPromptToOpenAI(prompt, model)
	case "Anthropic":
		response, err = sendPromptToAnthropic(prompt, model)
	case "Gemini":
		response, err = sendPromptToGemini(prompt, model)
	default:
		fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", provider)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Response:")
	fmt.Println(response)
}

func selectProvider() string {
	fmt.Println("Select provider:")
	fmt.Println("1) OpenAI")
	fmt.Println("2) Anthropic")
	fmt.Println("3) Gemini")
	fmt.Print("Enter choice: ")
	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid input: %v\nDefaulting to OpenAI\n", err)
		return "OpenAI"
	}
	switch choice {
	case 1:
		return "OpenAI"
	case 2:
		return "Anthropic"
	case 3:
		return "Gemini"
	default:
		fmt.Println("Invalid choice, defaulting to OpenAI")
		return "OpenAI"
	}
}

func selectModel(provider string) string {
	fmt.Printf("Enter the %s model you'd like to use: ", provider)
	var model string
	fmt.Scanln(&model)
	return strings.TrimSpace(model)
}

func readPrompt() string {
	fmt.Print("Enter your prompt: ")
	reader := bufio.NewReader(os.Stdin)
	prompt, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read prompt: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(prompt)
}

func sendPromptToOpenAI(prompt, model string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	reqBody := map[string]interface{}{"model": model, "messages": []map[string]string{{"role": "user", "content": prompt}}}
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

func sendPromptToAnthropic(prompt, model string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	anthropicPrompt := prompt
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": anthropicPrompt,
			},
		},
		"max_tokens":  5000,
		"temperature": 1.0,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error: %s", string(body))
	}
	var respBody struct {
		Content []struct {
			Role string `json:"role"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}
	return respBody.Content[0].Text, nil
}

func sendPromptToGemini(prompt, model string) (string, error) {
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

	result, err := client.Models.GenerateContent(
		context.Background(),
		model,
		genai.Text(prompt),
		nil,
	)

	return result.Text(), err
}
