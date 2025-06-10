package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Email EmailConfig `yaml:"email"`
}

type EmailConfig struct {
	SMTPHost       string `yaml:"smtp_host"`
	SMTPPort       int    `yaml:"smtp_port"`
	FromEmail      string `yaml:"from_email"`
	FromName       string `yaml:"from_name"`
	PasswordEnvVar string `yaml:"password_env_var"`
	SubjectPrefix  string `yaml:"subject_prefix"`
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
	config := &Config{
		Email: EmailConfig{
			SMTPHost:       "smtp.gmail.com",
			SMTPPort:       587,
			FromEmail:      "consensus.ai.25@gmail.com",
			FromName:       "Consensus AI",
			PasswordEnvVar: "CONSENSUS_EMAIL_PASSWORD",
			SubjectPrefix:  "[Consensus AI]",
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

