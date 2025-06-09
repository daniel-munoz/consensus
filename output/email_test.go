package output

import (
	"os"
	"strings"
	"testing"
)

func TestNewEmailOutput(t *testing.T) {
	// Test with environment variables set
	os.Setenv("CONSENSUS_EMAIL_PASSWORD", "test-password")
	os.Setenv("CONSENSUS_EMAIL_RECIPIENTS", "test1@example.com,test2@example.com")
	defer func() {
		os.Unsetenv("CONSENSUS_EMAIL_PASSWORD")
		os.Unsetenv("CONSENSUS_EMAIL_RECIPIENTS")
	}()

	email := NewEmailOutput()

	if email.FromEmail != "consensus.ai.25@gmail.com" {
		t.Errorf("Expected FromEmail to be consensus.ai.25@gmail.com, got %s", email.FromEmail)
	}

	if email.FromPassword != "test-password" {
		t.Errorf("Expected FromPassword to be test-password, got %s", email.FromPassword)
	}

	if len(email.ToEmails) != 2 {
		t.Errorf("Expected 2 recipient emails, got %d", len(email.ToEmails))
	}

	if email.ToEmails[0] != "test1@example.com" {
		t.Errorf("Expected first email to be test1@example.com, got %s", email.ToEmails[0])
	}

	if email.ToEmails[1] != "test2@example.com" {
		t.Errorf("Expected second email to be test2@example.com, got %s", email.ToEmails[1])
	}

	if email.SMTPHost != "smtp.gmail.com" {
		t.Errorf("Expected SMTPHost to be smtp.gmail.com, got %s", email.SMTPHost)
	}

	if email.SMTPPort != 587 {
		t.Errorf("Expected SMTPPort to be 587, got %d", email.SMTPPort)
	}
}

func TestEmailOutput_Name(t *testing.T) {
	email := NewEmailOutput()
	if email.Name() != "email" {
		t.Errorf("Expected Name() to return 'email', got %s", email.Name())
	}
}

func TestEmailOutput_WithIgnored(t *testing.T) {
	email := NewEmailOutput()
	result := email.WithIgnored("openai", "gemini")

	if result != email {
		t.Error("WithIgnored should return the same instance")
	}

	if _, exists := email.IgnoredProducers["openai"]; !exists {
		t.Error("openai should be in ignored producers")
	}

	if _, exists := email.IgnoredProducers["gemini"]; !exists {
		t.Error("gemini should be in ignored producers")
	}

	if _, exists := email.IgnoredProducers["anthropic"]; exists {
		t.Error("anthropic should not be in ignored producers")
	}
}

func TestEmailOutput_Send_WithIgnoredProducer(t *testing.T) {
	email := NewEmailOutput()
	email.WithIgnored("openai")

	// This should return without attempting to send
	err := email.Send("test content", "test-session", "openai")
	if err != nil {
		t.Errorf("Expected no error for ignored producer, got %v", err)
	}
}

func TestEmailOutput_Send_NoPasswordOrRecipients(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("CONSENSUS_EMAIL_PASSWORD")
	os.Unsetenv("CONSENSUS_EMAIL_RECIPIENTS")

	email := NewEmailOutput()

	// This should return without attempting to send
	err := email.Send("test content", "test-session", "openai")
	if err != nil {
		t.Errorf("Expected no error when no password/recipients configured, got %v", err)
	}
}

func TestEmailOutput_formatSubject(t *testing.T) {
	email := NewEmailOutput()
	sessionID := "test-session-123"

	tests := []struct {
		producer string
		expected string
	}{
		{"request", "[Consensus AI] Session test-session-123 - User Request"},
		{"prompt", "[Consensus AI] Session test-session-123 - Optimized Prompt"},
		{"openai", "[Consensus AI] Session test-session-123 - OpenAI Response"},
		{"anthropic", "[Consensus AI] Session test-session-123 - Anthropic Response"},
		{"gemini", "[Consensus AI] Session test-session-123 - Gemini Response"},
		{"unknown", "[Consensus AI] Session test-session-123 - Unknown"},
	}

	for _, test := range tests {
		result := email.formatSubject(sessionID, test.producer)
		if result != test.expected {
			t.Errorf("For producer %s, expected %s, got %s", test.producer, test.expected, result)
		}
	}
}

func TestEmailOutput_getContentType(t *testing.T) {
	email := NewEmailOutput()

	tests := []struct {
		producer string
		expected string
	}{
		{"request", "User Request"},
		{"prompt", "Optimized Prompt"},
		{"openai", "OpenAI Response"},
		{"anthropic", "Anthropic Response"},
		{"gemini", "Gemini Response"},
		{"unknown", "Unknown Response"},
	}

	for _, test := range tests {
		result := email.getContentType(test.producer)
		if result != test.expected {
			t.Errorf("For producer %s, expected %s, got %s", test.producer, test.expected, result)
		}
	}
}

func TestEmailOutput_getProviderBadge(t *testing.T) {
	email := NewEmailOutput()

	tests := []struct {
		producer string
		contains string
	}{
		{"openai", "openai"},
		{"anthropic", "anthropic"},
		{"gemini", "gemini"},
		{"request", "system"},
		{"prompt", "system"},
		{"unknown", ""},
	}

	for _, test := range tests {
		result := email.getProviderBadge(test.producer)
		if test.contains == "" {
			if result != "" {
				t.Errorf("For producer %s, expected empty badge, got %s", test.producer, result)
			}
		} else {
			if !strings.Contains(result, test.contains) {
				t.Errorf("For producer %s, expected badge to contain %s, got %s", test.producer, test.contains, result)
			}
		}
	}
}

func TestEmailOutput_formatBody(t *testing.T) {
	email := NewEmailOutput()
	content := "Test email content"
	sessionID := "test-session-123"
	producer := "openai"

	body := email.formatBody(content, sessionID, producer)

	// Check that the HTML contains expected elements
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Body should contain DOCTYPE declaration")
	}

	if !strings.Contains(body, sessionID) {
		t.Error("Body should contain session ID")
	}

	if !strings.Contains(body, content) {
		t.Error("Body should contain the content")
	}

	if !strings.Contains(body, "OpenAI Response") {
		t.Error("Body should contain content type")
	}

	if !strings.Contains(body, "openai") {
		t.Error("Body should contain provider badge")
	}
}