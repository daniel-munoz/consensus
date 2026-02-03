package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/daniel-munoz/consensus/ai"
)

type Config struct {
	Email            EmailConfig
	Providers        []Provider
	PromptProvider   Provider
	AllProviders     map[string]Provider
}

type ConfigYaml struct {
	Email             EmailConfig      `yaml:"email"`
	Providers         []ProviderConfig `yaml:"providers"`
	PromptProvider    string           `yaml:"prompt_provider"`
	ResponseProviders []string         `yaml:"response_providers"`
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
	Name           string  `yaml:"name"`
	Type           string  `yaml:"type"`
	BaseURL        *string `yaml:"base_url,omitempty"` // Optional, used for custom providers
	APIKeyVariable string  `yaml:"api_key_variable"`
	Model          string  `yaml:"model"`
	MaxTokens      *int64  `yaml:"max_tokens,omitempty"`
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	var configYaml ConfigYaml
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		createdConfig, err := createDefaultConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return yamlToConfig(createdConfig)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &configYaml); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return yamlToConfig(&configYaml)
}

func yamlToConfig(configYaml *ConfigYaml) (*Config, error) {
	providers := loadProviders(configYaml)
	providersByName := make(map[string]Provider)
	for _, provider := range providers {
		providersByName[provider.Name()] = provider
	}

	selectedProviders := make([]Provider, 0, len(configYaml.ResponseProviders))
	for _, name := range configYaml.ResponseProviders {
		if provider, exists := providersByName[name]; exists {
			selectedProviders = append(selectedProviders, provider)
		} else {
			return nil, fmt.Errorf("provider %s not found in config", name)
		}
	}

	if _, ok := providersByName[configYaml.PromptProvider]; !ok {
		return nil, fmt.Errorf("prompt provider %s not found in config", configYaml.PromptProvider)
	}

	config := &Config{
		Email:          configYaml.Email,
		Providers:      selectedProviders,
		PromptProvider: providersByName[configYaml.PromptProvider],
		AllProviders:   providersByName,
	}

	return config, nil
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

func createDefaultConfig(configPath string) (*ConfigYaml, error) {
	maxTokens := DefaultAnthropicMaxTokens
	config := &ConfigYaml{
		Email: EmailConfig{
			SMTPHost:       DefaultSMTPHost,
			SMTPPort:       DefaultSMTPPort,
			FromEmail:      DefaultFromEmail,
			FromName:       DefaultFromName,
			PasswordEnvVar: DefaultPasswordEnvVar,
			SubjectPrefix:  DefaultSubjectPrefix,
		},
		Providers: []ProviderConfig{
			{
				Name:           ProviderTypeOpenAI,
				Type:           ProviderTypeOpenAI,
				APIKeyVariable: "OPENAI_API_KEY",
				Model:          "gpt-4o",
			},
			{
				Name:           ProviderTypeGemini,
				Type:           ProviderTypeGemini,
				APIKeyVariable: "GEMINI_API_KEY",
				Model:          "gemini-2.0-flash",
			},
			{
				Name:           ProviderTypeAnthropic,
				Type:           ProviderTypeAnthropic,
				APIKeyVariable: "ANTHROPIC_API_KEY",
				Model:          "claude-4-sonnet-20250514",
				MaxTokens:      &maxTokens,
			},
		},
		PromptProvider:    ProviderTypeOpenAI,
		ResponseProviders: []string{ProviderTypeOpenAI, ProviderTypeGemini, ProviderTypeAnthropic},
	}

	if err := os.MkdirAll(filepath.Dir(configPath), ConfigDirPerms); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, data, ConfigFilePerms); err != nil {
		return nil, fmt.Errorf("failed to write default config: %w", err)
	}

	fmt.Printf("Created default config file at: %s\n", configPath)
	return config, nil
}

func loadProviders(config *ConfigYaml) []Provider {
	var providers []Provider

	for _, pc := range config.Providers {
		switch pc.Type {
		case ProviderTypeOpenAI:
			providers = append(providers, ai.NewOpenAI(pc.Name, pc.APIKeyVariable, pc.Model, pc.BaseURL))
		case ProviderTypeGemini:
			providers = append(providers, ai.NewGemini(pc.Name, pc.APIKeyVariable, pc.Model, pc.BaseURL))
		case ProviderTypeAnthropic:
			maxTokens := DefaultAnthropicMaxTokens
			if pc.MaxTokens != nil {
				maxTokens = *pc.MaxTokens
			}
			providers = append(providers, ai.NewAnthropic(pc.Name, pc.APIKeyVariable, pc.Model, maxTokens, pc.BaseURL))
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown provider type '%s' for provider '%s', skipping\n", pc.Type, pc.Name)
		}
	}

	return providers
}
