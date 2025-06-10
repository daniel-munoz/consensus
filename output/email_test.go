package output

import (
	"os"
	"strings"
	"testing"
)


func TestEmailOutput_Name(t *testing.T) {
	email := NewEmailOutputWithRecipients("")
	if email.Name() != "email" {
		t.Errorf("Expected Name() to return 'email', got %s", email.Name())
	}
}

func TestEmailOutput_WithIgnored(t *testing.T) {
	email := NewEmailOutputWithRecipients("")
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
	email := NewEmailOutputWithRecipients("")
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

	email := NewEmailOutputWithRecipients("")

	// This should return without attempting to send
	err := email.Send("test content", "test-session", "openai")
	if err != nil {
		t.Errorf("Expected no error when no password/recipients configured, got %v", err)
	}
}

func TestEmailOutput_formatSubject(t *testing.T) {
	email := NewEmailOutputWithRecipients("")
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
	email := NewEmailOutputWithRecipients("")

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
	email := NewEmailOutputWithRecipients("")

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
	email := NewEmailOutputWithRecipients("")
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

func TestNewEmailOutputWithRecipients(t *testing.T) {
	// Test with recipients provided
	recipients := "user1@example.com,user2@example.com, user3@example.com "
	
	// Set password env var for this test
	os.Setenv("CONSENSUS_EMAIL_PASSWORD", "test-password")
	defer os.Unsetenv("CONSENSUS_EMAIL_PASSWORD")

	email := NewEmailOutputWithRecipients(recipients)

	if email.FromEmail != "consensus.ai.25@gmail.com" {
		t.Errorf("Expected FromEmail to be consensus.ai.25@gmail.com, got %s", email.FromEmail)
	}

	if email.FromPassword != "test-password" {
		t.Errorf("Expected FromPassword to be test-password, got %s", email.FromPassword)
	}

	if len(email.ToEmails) != 3 {
		t.Errorf("Expected 3 recipient emails, got %d", len(email.ToEmails))
	}

	expectedEmails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for i, expected := range expectedEmails {
		if email.ToEmails[i] != expected {
			t.Errorf("Expected email %d to be %s, got %s", i, expected, email.ToEmails[i])
		}
	}

	if email.SMTPHost != "smtp.gmail.com" {
		t.Errorf("Expected SMTPHost to be smtp.gmail.com, got %s", email.SMTPHost)
	}

	if email.SMTPPort != 587 {
		t.Errorf("Expected SMTPPort to be 587, got %d", email.SMTPPort)
	}
}

func TestNewEmailOutputWithRecipients_EmptyString(t *testing.T) {
	// Test with empty recipients string
	email := NewEmailOutputWithRecipients("")

	if len(email.ToEmails) != 0 {
		t.Errorf("Expected 0 recipient emails for empty string, got %d", len(email.ToEmails))
	}
}

func TestNewEmailOutputWithRecipients_WithEmptyEntries(t *testing.T) {
	// Test with recipients that have empty entries
	recipients := "user1@example.com,,user2@example.com,  ,user3@example.com"
	
	email := NewEmailOutputWithRecipients(recipients)

	if len(email.ToEmails) != 3 {
		t.Errorf("Expected 3 recipient emails after filtering, got %d", len(email.ToEmails))
	}

	expectedEmails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for i, expected := range expectedEmails {
		if email.ToEmails[i] != expected {
			t.Errorf("Expected email %d to be %s, got %s", i, expected, email.ToEmails[i])
		}
	}
}