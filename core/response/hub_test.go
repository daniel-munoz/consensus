package response

import (
	"errors"
	"testing"
)

func TestNewHub(t *testing.T) {
	promptProvider := "openai"
	providers := []string{"openai", "anthropic", "gemini"}
	promptResponse := DelayedResponse{}
	responses := make(map[string]DelayedResponse)

	hub := NewHub(promptProvider, providers, promptResponse, responses)

	if hub == nil {
		t.Error("Expected hub to be non-nil")
	}

	if hub.MasterPromptPoviderName() != promptProvider {
		t.Errorf("Expected provider name '%s', got '%s'", promptProvider, hub.MasterPromptPoviderName())
	}

	hubProviders := hub.Providers()
	if len(hubProviders) != len(providers) {
		t.Errorf("Expected %d providers, got %d", len(providers), len(hubProviders))
	}

	for i, provider := range providers {
		if hubProviders[i] != provider {
			t.Errorf("Expected provider '%s' at index %d, got '%s'", provider, i, hubProviders[i])
		}
	}
}

func TestHub_MasterPromptPoviderName(t *testing.T) {
	promptProvider := "anthropic"
	hub := NewHub(promptProvider, []string{}, DelayedResponse{}, make(map[string]DelayedResponse))

	result := hub.MasterPromptPoviderName()
	if result != promptProvider {
		t.Errorf("Expected '%s', got '%s'", promptProvider, result)
	}
}

func TestHub_MasterPrompt_Success(t *testing.T) {
	valueQueue := make(chan string, 1)
	valueQueue <- "optimized prompt"
	
	promptResponse := DelayedResponse{
		timeoutInMinutes: 1,
		valueQueue:       valueQueue,
		errorQueue:       make(chan error, 1),
		responded:        false,
	}

	hub := NewHub("openai", []string{}, promptResponse, make(map[string]DelayedResponse))

	result, err := hub.MasterPrompt()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "optimized prompt" {
		t.Errorf("Expected 'optimized prompt', got '%s'", result)
	}
}

func TestHub_MasterPrompt_Error(t *testing.T) {
	errorQueue := make(chan error, 1)
	expectedErr := errors.New("prompt generation failed")
	errorQueue <- expectedErr
	
	promptResponse := DelayedResponse{
		timeoutInMinutes: 1,
		valueQueue:       make(chan string, 1),
		errorQueue:       errorQueue,
		responded:        false,
	}

	hub := NewHub("openai", []string{}, promptResponse, make(map[string]DelayedResponse))

	result, err := hub.MasterPrompt()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestHub_Providers(t *testing.T) {
	providers := []string{"openai", "anthropic", "gemini"}
	hub := NewHub("openai", providers, DelayedResponse{}, make(map[string]DelayedResponse))

	result := hub.Providers()
	if len(result) != len(providers) {
		t.Errorf("Expected %d providers, got %d", len(providers), len(result))
	}

	for i, provider := range providers {
		if result[i] != provider {
			t.Errorf("Expected provider '%s' at index %d, got '%s'", provider, i, result[i])
		}
	}
}

func TestHub_ResponseFrom_Success(t *testing.T) {
	valueQueue := make(chan string, 1)
	valueQueue <- "anthropic response"
	
	anthropicResponse := DelayedResponse{
		timeoutInMinutes: 1,
		valueQueue:       valueQueue,
		errorQueue:       make(chan error, 1),
		responded:        false,
	}

	responses := map[string]DelayedResponse{
		"anthropic": anthropicResponse,
	}

	hub := NewHub("openai", []string{"anthropic"}, DelayedResponse{}, responses)

	result, err := hub.ResponseFrom("anthropic")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "anthropic response" {
		t.Errorf("Expected 'anthropic response', got '%s'", result)
	}
}

func TestHub_ResponseFrom_ProviderNotFound(t *testing.T) {
	responses := map[string]DelayedResponse{}
	hub := NewHub("openai", []string{}, DelayedResponse{}, responses)

	result, err := hub.ResponseFrom("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent provider, got nil")
	}
	if err.Error() != "provider not found: nonexistent" {
		t.Errorf("Expected 'provider not found: nonexistent', got '%s'", err.Error())
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestHub_ResponseFrom_ProviderError(t *testing.T) {
	errorQueue := make(chan error, 1)
	expectedErr := errors.New("anthropic API error")
	errorQueue <- expectedErr
	
	anthropicResponse := DelayedResponse{
		timeoutInMinutes: 1,
		valueQueue:       make(chan string, 1),
		errorQueue:       errorQueue,
		responded:        false,
	}

	responses := map[string]DelayedResponse{
		"anthropic": anthropicResponse,
	}

	hub := NewHub("openai", []string{"anthropic"}, DelayedResponse{}, responses)

	result, err := hub.ResponseFrom("anthropic")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}