package main

import (
	"os"
	"path/filepath"
	"testing"
	"gopkg.in/yaml.v3"
)

func TestLoadProvidersFromConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	
	maxTokens := int64(32000)
	testConfig := &Config{
		Email: EmailConfig{
			SMTPHost:       "smtp.example.com",
			SMTPPort:       587,
			FromEmail:      "test@example.com",
			FromName:       "Test",
			PasswordEnvVar: "TEST_PASSWORD",
			SubjectPrefix:  "[Test]",
		},
		Providers: []ProviderConfig{
			{
				Name:           "test-openai",
				Type:           "openai",
				APIKeyVariable: "TEST_OPENAI_KEY",
				Model:          "gpt-3.5-turbo",
			},
			{
				Name:           "test-anthropic",
				Type:           "anthropic",
				APIKeyVariable: "TEST_ANTHROPIC_KEY",
				Model:          "claude-3-sonnet",
				MaxTokens:      &maxTokens,
			},
		},
	}
	
	// Write test config to file
	data, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	
	// Load providers from config
	providers := LoadProviders(testConfig)
	
	// Verify we got the expected number of providers
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}
	
	// Verify provider names
	providerNames := make(map[string]bool)
	for _, provider := range providers {
		providerNames[provider.Name()] = true
	}
	
	if !providerNames["test-openai"] {
		t.Error("Expected test-openai provider not found")
	}
	
	if !providerNames["test-anthropic"] {
		t.Error("Expected test-anthropic provider not found")
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	
	config, err := createDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}
	
	// Verify default providers are created
	if len(config.Providers) != 3 {
		t.Errorf("Expected 3 default providers, got %d", len(config.Providers))
	}
	
	// Verify provider types
	providerTypes := make(map[string]bool)
	for _, provider := range config.Providers {
		providerTypes[provider.Type] = true
	}
	
	expectedTypes := []string{"openai", "gemini", "anthropic"}
	for _, expectedType := range expectedTypes {
		if !providerTypes[expectedType] {
			t.Errorf("Expected provider type %s not found", expectedType)
		}
	}
	
	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}