package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var inputReader io.Reader = os.Stdin

const versionNumber = "1.0.3"

type Response struct {
	Text         string
	Provider     string
	ProviderType string
}

type Provider interface {
	Name() string
	Type() string
	Send(string, *string) (string, error)
}

func readPrompt() string {
	fmt.Print("Enter your prompt: ")
	reader := bufio.NewReader(inputReader)
	prompt, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read prompt: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(prompt)
}

func overrideConfigWithFlags(config *Config, masterPromptProvider, responseProviders string) error {
	if masterPromptProvider != "" {
		if provider, exists := config.AllProviders[masterPromptProvider]; exists {
			config.PromptProvider = provider
		} else {
			return fmt.Errorf("master prompt provider '%s' not found in available providers", masterPromptProvider)
		}
	}

	if responseProviders != "" {
		providerNames := strings.Split(responseProviders, ",")
		var selectedProviders []Provider

		for _, name := range providerNames {
			name = strings.TrimSpace(name)
			if provider, exists := config.AllProviders[name]; exists {
				selectedProviders = append(selectedProviders, provider)
			} else {
				return fmt.Errorf("response provider '%s' not found in available providers", name)
			}
		}

		config.Providers = selectedProviders
	}

	return nil
}

func main() {
	flags := ParseFlags()

	if flags.Version {
		fmt.Printf("consensus version %s\n", versionNumber)
		return
	}

	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Server mode: start HTTP server with web UI
	if flags.Serve {
		server := NewServer(config, flags.Port, flags.MultiSession)

		// Handle graceful shutdown on interrupt signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\nReceived interrupt signal, shutting down...")
			server.Shutdown()
		}()

		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI mode: process prompt directly
	if err := overrideConfigWithFlags(config, flags.MasterPromptProvider, flags.ResponseProviders); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to override config: %v\n", err)
		os.Exit(1)
	}

	request := flags.Prompt
	if request == "" {
		request = readPrompt()
	}

	sessionID := NewSession()
	outputManager := NewOutputManager(config, flags.EmailTo)

	outputManager.Send(request, sessionID, "request", "request")

	prompt := request
	if !flags.NoMasterPrompt {
		prompt, err = OptimizePrompt(request, config.PromptProvider, sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputManager.Send(prompt, sessionID, "prompt", "prompt")
	}

	runProvidersConcurrently(config.Providers, prompt, sessionID, outputManager)
}
