package ai

import (
	"context"
	"errors"
	"os"
	"testing"
)

// MockOpenAIClient is a mock implementation of OpenAIClient for testing
type MockOpenAIClient struct {
	response   string
	err        error
	callCount  int
	lastModel  string
	lastInput  string
	lastSystem *string
}

func (m *MockOpenAIClient) CreateResponse(ctx context.Context, model, input string, instructions *string) (string, error) {
	m.callCount++
	m.lastModel = model
	m.lastInput = input
	m.lastSystem = instructions
	return m.response, m.err
}

func (m *MockOpenAIClient) SetResponse(response string) { m.response = response }
func (m *MockOpenAIClient) SetError(err error)          { m.err = err }

// MockAnthropicClient is a mock implementation of AnthropicClient for testing
type MockAnthropicClient struct {
	response     string
	err          error
	callCount    int
	lastModel    string
	lastMaxTokens int64
	lastPrompt   string
	lastSystem   *string
}

func (m *MockAnthropicClient) CreateMessage(ctx context.Context, model string, maxTokens int64, prompt string, system *string) (string, error) {
	m.callCount++
	m.lastModel = model
	m.lastMaxTokens = maxTokens
	m.lastPrompt = prompt
	m.lastSystem = system
	return m.response, m.err
}

func (m *MockAnthropicClient) SetResponse(response string) { m.response = response }
func (m *MockAnthropicClient) SetError(err error)          { m.err = err }

// MockGeminiClient is a mock implementation of GeminiClient for testing
type MockGeminiClient struct {
	response   string
	err        error
	callCount  int
	lastModel  string
	lastPrompt string
	lastSystem *string
}

func (m *MockGeminiClient) GenerateContent(ctx context.Context, model, prompt string, system *string) (string, error) {
	m.callCount++
	m.lastModel = model
	m.lastPrompt = prompt
	m.lastSystem = system
	return m.response, m.err
}

func (m *MockGeminiClient) SetResponse(response string) { m.response = response }
func (m *MockGeminiClient) SetError(err error)          { m.err = err }

// =============================================================================
// OpenAI Tests
// =============================================================================

func TestNewOpenAI(t *testing.T) {
	baseURL := "https://custom.openai.com"
	provider := NewOpenAI("test-openai", "TEST_API_KEY", "gpt-4", &baseURL)

	if provider.ConfigName != "test-openai" {
		t.Errorf("expected ConfigName 'test-openai', got '%s'", provider.ConfigName)
	}
	if provider.APIKeyVariable != "TEST_API_KEY" {
		t.Errorf("expected APIKeyVariable 'TEST_API_KEY', got '%s'", provider.APIKeyVariable)
	}
	if provider.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got '%s'", provider.Model)
	}
	if provider.BaseURL == nil || *provider.BaseURL != baseURL {
		t.Errorf("expected BaseURL '%s', got '%v'", baseURL, provider.BaseURL)
	}
}

func TestOpenAI_Name(t *testing.T) {
	provider := NewOpenAI("my-openai", "API_KEY", "gpt-4", nil)
	if provider.Name() != "my-openai" {
		t.Errorf("expected Name() to return 'my-openai', got '%s'", provider.Name())
	}
}

func TestOpenAI_Type(t *testing.T) {
	provider := NewOpenAI("test", "API_KEY", "gpt-4", nil)
	if provider.Type() != "openai" {
		t.Errorf("expected Type() to return 'openai', got '%s'", provider.Type())
	}
}

func TestOpenAI_Send_MissingAPIKey(t *testing.T) {
	provider := NewOpenAI("test", "NONEXISTENT_API_KEY_VAR", "gpt-4", nil)

	os.Unsetenv("NONEXISTENT_API_KEY_VAR")

	_, err := provider.Send("test prompt", nil)
	if err == nil {
		t.Error("expected error when API key is not set")
	}
	if err.Error() != "NONEXISTENT_API_KEY_VAR environment variable not set" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOpenAI_Send_Success(t *testing.T) {
	mock := &MockOpenAIClient{response: "Hello from OpenAI!"}

	provider := NewOpenAIWithClient(
		"test-openai",
		"TEST_OPENAI_API_KEY",
		"gpt-4",
		nil,
		func(apiKey string, baseURL *string) OpenAIClient { return mock },
	)

	os.Setenv("TEST_OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("TEST_OPENAI_API_KEY")

	result, err := provider.Send("test prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from OpenAI!" {
		t.Errorf("expected 'Hello from OpenAI!', got '%s'", result)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
	if mock.lastModel != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", mock.lastModel)
	}
	if mock.lastInput != "test prompt" {
		t.Errorf("expected input 'test prompt', got '%s'", mock.lastInput)
	}
}

func TestOpenAI_Send_WithSystemPrompt(t *testing.T) {
	mock := &MockOpenAIClient{response: "Response with system"}

	provider := NewOpenAIWithClient(
		"test-openai",
		"TEST_OPENAI_API_KEY",
		"gpt-4",
		nil,
		func(apiKey string, baseURL *string) OpenAIClient { return mock },
	)

	os.Setenv("TEST_OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("TEST_OPENAI_API_KEY")

	system := "You are a helpful assistant"
	_, err := provider.Send("test prompt", &system)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastSystem == nil || *mock.lastSystem != system {
		t.Errorf("expected system '%s', got '%v'", system, mock.lastSystem)
	}
}

func TestOpenAI_Send_ClientError(t *testing.T) {
	expectedErr := errors.New("API error")
	mock := &MockOpenAIClient{err: expectedErr}

	provider := NewOpenAIWithClient(
		"test-openai",
		"TEST_OPENAI_API_KEY",
		"gpt-4",
		nil,
		func(apiKey string, baseURL *string) OpenAIClient { return mock },
	)

	os.Setenv("TEST_OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("TEST_OPENAI_API_KEY")

	_, err := provider.Send("test prompt", nil)
	if err == nil {
		t.Error("expected error from client")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

// =============================================================================
// Anthropic Tests
// =============================================================================

func TestNewAnthropic(t *testing.T) {
	baseURL := "https://custom.anthropic.com"
	provider := NewAnthropic("test-anthropic", "TEST_API_KEY", "claude-3", 1000, &baseURL)

	if provider.ConfigName != "test-anthropic" {
		t.Errorf("expected ConfigName 'test-anthropic', got '%s'", provider.ConfigName)
	}
	if provider.APIKeyVariable != "TEST_API_KEY" {
		t.Errorf("expected APIKeyVariable 'TEST_API_KEY', got '%s'", provider.APIKeyVariable)
	}
	if provider.Model != "claude-3" {
		t.Errorf("expected Model 'claude-3', got '%s'", provider.Model)
	}
	if provider.MaxTokens != 1000 {
		t.Errorf("expected MaxTokens 1000, got %d", provider.MaxTokens)
	}
	if provider.BaseURL == nil || *provider.BaseURL != baseURL {
		t.Errorf("expected BaseURL '%s', got '%v'", baseURL, provider.BaseURL)
	}
}

func TestAnthropic_Name(t *testing.T) {
	provider := NewAnthropic("my-anthropic", "API_KEY", "claude-3", 1000, nil)
	if provider.Name() != "my-anthropic" {
		t.Errorf("expected Name() to return 'my-anthropic', got '%s'", provider.Name())
	}
}

func TestAnthropic_Type(t *testing.T) {
	provider := NewAnthropic("test", "API_KEY", "claude-3", 1000, nil)
	if provider.Type() != "anthropic" {
		t.Errorf("expected Type() to return 'anthropic', got '%s'", provider.Type())
	}
}

func TestAnthropic_Send_MissingAPIKey(t *testing.T) {
	provider := NewAnthropic("test", "NONEXISTENT_ANTHROPIC_KEY", "claude-3", 1000, nil)

	os.Unsetenv("NONEXISTENT_ANTHROPIC_KEY")

	_, err := provider.Send("test prompt", nil)
	if err == nil {
		t.Error("expected error when API key is not set")
	}
	if err.Error() != "NONEXISTENT_ANTHROPIC_KEY environment variable not set" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAnthropic_Send_Success(t *testing.T) {
	mock := &MockAnthropicClient{response: "Hello from Anthropic!"}

	provider := NewAnthropicWithClient(
		"test-anthropic",
		"TEST_ANTHROPIC_API_KEY",
		"claude-3",
		2000,
		nil,
		func(apiKey string, baseURL *string) AnthropicClient { return mock },
	)

	os.Setenv("TEST_ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("TEST_ANTHROPIC_API_KEY")

	result, err := provider.Send("test prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from Anthropic!" {
		t.Errorf("expected 'Hello from Anthropic!', got '%s'", result)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
	if mock.lastModel != "claude-3" {
		t.Errorf("expected model 'claude-3', got '%s'", mock.lastModel)
	}
	if mock.lastMaxTokens != 2000 {
		t.Errorf("expected maxTokens 2000, got %d", mock.lastMaxTokens)
	}
	if mock.lastPrompt != "test prompt" {
		t.Errorf("expected prompt 'test prompt', got '%s'", mock.lastPrompt)
	}
}

func TestAnthropic_Send_WithSystemPrompt(t *testing.T) {
	mock := &MockAnthropicClient{response: "Response with system"}

	provider := NewAnthropicWithClient(
		"test-anthropic",
		"TEST_ANTHROPIC_API_KEY",
		"claude-3",
		1000,
		nil,
		func(apiKey string, baseURL *string) AnthropicClient { return mock },
	)

	os.Setenv("TEST_ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("TEST_ANTHROPIC_API_KEY")

	system := "You are a helpful assistant"
	_, err := provider.Send("test prompt", &system)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastSystem == nil || *mock.lastSystem != system {
		t.Errorf("expected system '%s', got '%v'", system, mock.lastSystem)
	}
}

func TestAnthropic_Send_ClientError(t *testing.T) {
	expectedErr := errors.New("Anthropic API error")
	mock := &MockAnthropicClient{err: expectedErr}

	provider := NewAnthropicWithClient(
		"test-anthropic",
		"TEST_ANTHROPIC_API_KEY",
		"claude-3",
		1000,
		nil,
		func(apiKey string, baseURL *string) AnthropicClient { return mock },
	)

	os.Setenv("TEST_ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("TEST_ANTHROPIC_API_KEY")

	_, err := provider.Send("test prompt", nil)
	if err == nil {
		t.Error("expected error from client")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

// =============================================================================
// Gemini Tests
// =============================================================================

func TestNewGemini(t *testing.T) {
	baseURL := "https://custom.gemini.com"
	provider := NewGemini("test-gemini", "TEST_API_KEY", "gemini-pro", &baseURL)

	if provider.ConfigName != "test-gemini" {
		t.Errorf("expected ConfigName 'test-gemini', got '%s'", provider.ConfigName)
	}
	if provider.APIKeyVariable != "TEST_API_KEY" {
		t.Errorf("expected APIKeyVariable 'TEST_API_KEY', got '%s'", provider.APIKeyVariable)
	}
	if provider.Model != "gemini-pro" {
		t.Errorf("expected Model 'gemini-pro', got '%s'", provider.Model)
	}
	if provider.BaseURL == nil || *provider.BaseURL != baseURL {
		t.Errorf("expected BaseURL '%s', got '%v'", baseURL, provider.BaseURL)
	}
}

func TestGemini_Name(t *testing.T) {
	provider := NewGemini("my-gemini", "API_KEY", "gemini-pro", nil)
	if provider.Name() != "my-gemini" {
		t.Errorf("expected Name() to return 'my-gemini', got '%s'", provider.Name())
	}
}

func TestGemini_Type(t *testing.T) {
	provider := NewGemini("test", "API_KEY", "gemini-pro", nil)
	if provider.Type() != "gemini" {
		t.Errorf("expected Type() to return 'gemini', got '%s'", provider.Type())
	}
}

func TestGemini_Send_MissingAPIKey(t *testing.T) {
	provider := NewGemini("test", "NONEXISTENT_GEMINI_KEY", "gemini-pro", nil)

	os.Unsetenv("NONEXISTENT_GEMINI_KEY")

	_, err := provider.Send("test prompt", nil)
	if err == nil {
		t.Error("expected error when API key is not set")
	}
	if err.Error() != "NONEXISTENT_GEMINI_KEY environment variable not set" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGemini_Send_Success(t *testing.T) {
	mock := &MockGeminiClient{response: "Hello from Gemini!"}

	provider := NewGeminiWithClient(
		"test-gemini",
		"TEST_GEMINI_API_KEY",
		"gemini-pro",
		nil,
		func(apiKey string, baseURL *string) GeminiClient { return mock },
	)

	os.Setenv("TEST_GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("TEST_GEMINI_API_KEY")

	result, err := provider.Send("test prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from Gemini!" {
		t.Errorf("expected 'Hello from Gemini!', got '%s'", result)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
	if mock.lastModel != "gemini-pro" {
		t.Errorf("expected model 'gemini-pro', got '%s'", mock.lastModel)
	}
	if mock.lastPrompt != "test prompt" {
		t.Errorf("expected prompt 'test prompt', got '%s'", mock.lastPrompt)
	}
}

func TestGemini_Send_WithSystemPrompt(t *testing.T) {
	mock := &MockGeminiClient{response: "Response with system"}

	provider := NewGeminiWithClient(
		"test-gemini",
		"TEST_GEMINI_API_KEY",
		"gemini-pro",
		nil,
		func(apiKey string, baseURL *string) GeminiClient { return mock },
	)

	os.Setenv("TEST_GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("TEST_GEMINI_API_KEY")

	system := "You are a helpful assistant"
	_, err := provider.Send("test prompt", &system)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastSystem == nil || *mock.lastSystem != system {
		t.Errorf("expected system '%s', got '%v'", system, mock.lastSystem)
	}
}

func TestGemini_Send_ClientError(t *testing.T) {
	expectedErr := errors.New("Gemini API error")
	mock := &MockGeminiClient{err: expectedErr}

	provider := NewGeminiWithClient(
		"test-gemini",
		"TEST_GEMINI_API_KEY",
		"gemini-pro",
		nil,
		func(apiKey string, baseURL *string) GeminiClient { return mock },
	)

	os.Setenv("TEST_GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("TEST_GEMINI_API_KEY")

	_, err := provider.Send("test prompt", nil)
	if err == nil {
		t.Error("expected error from client")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}
