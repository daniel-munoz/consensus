package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/daniel-munoz/consensus/output"
	"github.com/google/uuid"
)

var inputReader io.Reader = os.Stdin

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
	// Override master prompt provider if specified
	if masterPromptProvider != "" {
		if provider, exists := config.AllProviders[masterPromptProvider]; exists {
			config.PromptProvider = provider
		} else {
			return fmt.Errorf("master prompt provider '%s' not found in available providers", masterPromptProvider)
		}
	}

	// Override response providers if specified
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
	promptFlag := flag.String("prompt", "", "Prompt text to use instead of interactive input")
	flag.StringVar(promptFlag, "p", "", "Prompt text to use instead of interactive input (shorthand)")
	emailToFlag := flag.String("email-to", "", "Comma-separated list of email addresses to send notifications to")
	flag.StringVar(emailToFlag, "e", "", "Comma-separated list of email addresses to send notifications to (shorthand)")
	noMasterPrompt := flag.Bool("no-master-prompt", false, "Disable master prompt for optimization")
	flag.BoolVar(noMasterPrompt, "nmp", false, "Disable master prompt for optimization (shorthand)")
	masterPromptProvider := flag.String("master-prompt-provider", "", "Provider to use for master prompt optimization")
	flag.StringVar(masterPromptProvider, "mpp", "", "Provider to use for master prompt optimization (shorthand)")
	responseProviders := flag.String("response-providers", "", "Comma-separated list of providers to use for responses")
	flag.StringVar(responseProviders, "rp", "", "Comma-separated list of providers to use for responses (shorthand)")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override config with command line flags
	if err := overrideConfigWithFlags(config, *masterPromptProvider, *responseProviders); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to override config: %v\n", err)
		os.Exit(1)
	}

	var (
		request   string
		responses = make(chan Response)
		done      = make(chan struct{})
	)

	if promptFlag != nil && *promptFlag != "" {
		request = *promptFlag
	} else {
		request = readPrompt()
	}

	// Generate UUID for this session
	id := uuid.NewString()

	// Initialize output manager with file and email output
	emailConfig := output.EmailConfig{
		SMTPHost:       config.Email.SMTPHost,
		SMTPPort:       config.Email.SMTPPort,
		FromEmail:      config.Email.FromEmail,
		FromName:       config.Email.FromName,
		PasswordEnvVar: config.Email.PasswordEnvVar,
		SubjectPrefix:  config.Email.SubjectPrefix,
	}
	outputManager := output.NewManager(
		output.NewFileOutput("responses"),
		output.NewEmailOutputWithRecipients(emailConfig, *emailToFlag),
	)

	outputManager.Send(request, id, "request", "request")

	var prompt string

	if !*noMasterPrompt {
		// Use the master prompt for optimization
		developerInstructions := masterPrompt

		fmt.Printf("Creating final prompt for request %s\n", id)

		requestToPromptText := fmt.Sprintf("Create the best prompt to address the following request: %s", request)

		promptProvider := config.PromptProvider

		prompt, err = promptProvider.Send(requestToPromptText, &developerInstructions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputManager.Send(prompt, id, "prompt", "prompt")
	} else {
		// Use the request directly as the prompt
		prompt = request
	}

	waitGroup := sync.WaitGroup{}

	for _, provider := range config.Providers {
		waitGroup.Add(1)
		go func(p Provider) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "Panic from %s: %v\n", p.Name(), r)
				}
				waitGroup.Done()
			}()
			fmt.Printf("Consulting %s...\n", p.Name())
			response, err := p.Send(prompt, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from %s: %v\n", p.Name(), err)
				return
			}
			responses <- Response{Text: response, Provider: p.Name(), ProviderType: p.Type()}
			fmt.Printf("%s responded!\n", p.Name())
		}(provider)
	}

	go func() {
		for response := range responses {
			outputManager.Send(response.Text, id, response.Provider, response.ProviderType)
		}
		done <- struct{}{}
	}()

	waitGroup.Wait()
	close(responses)
	<-done
}
