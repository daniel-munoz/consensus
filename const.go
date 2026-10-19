package main

// Provider defaults
const (
	DefaultAnthropicMaxTokens int64 = 64000
	DefaultSMTPPort                 = 587
	DefaultSMTPHost                 = "smtp.gmail.com"
	DefaultFromEmail                = "consensus.ai.25@gmail.com"
	DefaultFromName                 = "Consensus AI"
	DefaultPasswordEnvVar           = "CONSENSUS_EMAIL_PASSWORD"
	DefaultSubjectPrefix            = "[Consensus AI]"
)

// File permissions
const (
	ConfigDirPerms  = 0755
	ConfigFilePerms = 0644
)

// Provider type identifiers
const (
	ProviderTypeOpenAI    = "openai"
	ProviderTypeAnthropic = "anthropic"
	ProviderTypeGemini    = "gemini"
)
