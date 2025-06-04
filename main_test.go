package main

import (
	"os"
	"strings"
	"testing"
)

func TestSelectProvider_ValidChoices(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1\n", "OpenAI"},
		{"2\n", "Anthropic"},
		{"3\n", "Gemini"},
	}
	for _, tt := range tests {
		oldInput := inputReader
		inputReader = strings.NewReader(tt.input)
		got := selectProvider()
		inputReader = oldInput
		if got != tt.expected {
			t.Errorf("selectProvider() = %q; want %q", got, tt.expected)
		}
	}
}

func TestSelectProvider_InvalidInput_DefaultsToOpenAI(t *testing.T) {
	oldInput := inputReader
	inputReader = strings.NewReader("invalid\n")
	got := selectProvider()
	inputReader = oldInput
	if got != "OpenAI" {
		t.Errorf("selectProvider() for invalid input = %q; want %q", got, "OpenAI")
	}
}

func TestSelectProvider_InvalidChoice_DefaultsToOpenAI(t *testing.T) {
	oldInput := inputReader
	inputReader = strings.NewReader("99\n")
	got := selectProvider()
	inputReader = oldInput
	if got != "OpenAI" {
		t.Errorf("selectProvider() for out-of-range choice = %q; want %q", got, "OpenAI")
	}
}

func TestSelectModel_TrimsSpaces(t *testing.T) {
	oldInput := inputReader
	inputReader = strings.NewReader("  my-model \n")
	got := selectModel("TestProvider")
	inputReader = oldInput
	if got != "my-model" {
		t.Errorf("selectModel() = %q; want %q", got, "my-model")
	}
}

func TestReadPrompt_TrimsNewline(t *testing.T) {
	oldInput := inputReader
	inputReader = strings.NewReader("hello prompt\n")
	got := readPrompt()
	inputReader = oldInput
	if got != "hello prompt" {
		t.Errorf("readPrompt() = %q; want %q", got, "hello prompt")
	}
}

func TestSendPromptToOpenAI_MissingAPIKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	_, err := sendPromptToOpenAI("p", "m")
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY environment variable not set") {
		t.Errorf("sendPromptToOpenAI without API key error = %v; want missing key error", err)
	}
}

func TestSendPromptToAnthropic_MissingAPIKey(t *testing.T) {
	os.Unsetenv("ANTHROPIC_API_KEY")
	_, err := sendPromptToAnthropic("p", "m")
	if err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY environment variable not set") {
		t.Errorf("sendPromptToAnthropic without API key error = %v; want missing key error", err)
	}
}

func TestSendPromptToGemini_MissingAPIKey(t *testing.T) {
	os.Unsetenv("GEMINI_API_KEY")
	_, err := sendPromptToGemini("p", "m")
	if err == nil || !strings.Contains(err.Error(), "GEMINI_API_KEY environment variable not set") {
		t.Errorf("sendPromptToGemini without API key error = %v; want missing key error", err)
	}
}
