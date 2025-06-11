package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/daniel-munoz/consensus/ai"
)

type Config struct {
	Email     EmailConfig      `yaml:"email"`
	Providers []ProviderConfig `yaml:"providers"`
}

type EmailConfig struct {
	SMTPHost       string `yaml:"smtp_host"`
	SMTPPort       int    `yaml:"smtp_port"`
	FromEmail      string `yaml:"from_email"`
	FromName       string `yaml:"from_name"`
	PasswordEnvVar string `yaml:"password_env_var"`
	SubjectPrefix  string `yaml:"subject_prefix"`
}

type ProviderConfig struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
	APIKeyVariable string `yaml:"api_key_variable"`
	Model          string `yaml:"model"`
	MaxTokens      *int64 `yaml:"max_tokens,omitempty"`
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return createDefaultConfig(configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func getConfigPath() string {
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		return filepath.Join(configDir, "consensus", "config.yml")
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".config", "consensus", "config.yml")
	}

	return "config.yml"
}

func createDefaultConfig(configPath string) (*Config, error) {
	maxTokens := int64(64000)
	config := &Config{
		Email: EmailConfig{
			SMTPHost:       "smtp.gmail.com",
			SMTPPort:       587,
			FromEmail:      "consensus.ai.25@gmail.com",
			FromName:       "Consensus AI",
			PasswordEnvVar: "CONSENSUS_EMAIL_PASSWORD",
			SubjectPrefix:  "[Consensus AI]",
		},
		Providers: []ProviderConfig{
			{
				Name:           "openai",
				Type:           "openai",
				APIKeyVariable: "OPENAI_API_KEY",
				Model:          "gpt-4o",
			},
			{
				Name:           "gemini",
				Type:           "gemini",
				APIKeyVariable: "GEMINI_API_KEY",
				Model:          "gemini-2.0-flash",
			},
			{
				Name:           "anthropic",
				Type:           "anthropic",
				APIKeyVariable: "ANTHROPIC_API_KEY",
				Model:          "claude-4-sonnet-20250514",
				MaxTokens:      &maxTokens,
			},
		},
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write default config: %w", err)
	}

	fmt.Printf("Created default config file at: %s\n", configPath)
	return config, nil
}

func LoadProviders(config *Config) []Provider {
	var providers []Provider
	
	for _, pc := range config.Providers {
		switch pc.Type {
		case "openai":
			providers = append(providers, ai.NewOpenAI(pc.Name, pc.APIKeyVariable, pc.Model))
		case "gemini":
			providers = append(providers, ai.NewGemini(pc.Name, pc.APIKeyVariable, pc.Model))
		case "anthropic":
			maxTokens := int64(64000)
			if pc.MaxTokens != nil {
				maxTokens = *pc.MaxTokens
			}
			providers = append(providers, ai.NewAnthropic(pc.Name, pc.APIKeyVariable, pc.Model, maxTokens))
		}
	}
	
	return providers
}
