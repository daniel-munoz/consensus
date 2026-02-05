package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/daniel-munoz/consensus/output"
	"github.com/google/uuid"
)

// CLIFlags holds all command-line flag values
type CLIFlags struct {
	Prompt               string
	EmailTo              string
	NoMasterPrompt       bool
	MasterPromptProvider string
	ResponseProviders    string
	Version              bool
	Serve                bool
	Port                 int
	MultiSession         bool
}

// ParseFlags parses command-line flags and returns a CLIFlags struct
func ParseFlags() CLIFlags {
	var flags CLIFlags

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
	versionFlag := flag.Bool("version", false, "Print version and exit")
	serveFlag := flag.Bool("serve", false, "Start HTTP server mode with web UI")
	portFlag := flag.Int("port", 8080, "HTTP server port (used with --serve)")
	multiSessionFlag := flag.Bool("multi-session", false, "Keep server running for multiple requests (used with --serve)")
	flag.Parse()

	flags.Prompt = *promptFlag
	flags.EmailTo = *emailToFlag
	flags.NoMasterPrompt = *noMasterPrompt
	flags.MasterPromptProvider = *masterPromptProvider
	flags.ResponseProviders = *responseProviders
	flags.Version = *versionFlag
	flags.Serve = *serveFlag
	flags.Port = *portFlag
	flags.MultiSession = *multiSessionFlag

	return flags
}

// NewOutputManager creates the output manager with file and email outputs
func NewOutputManager(config *Config, emailTo string) *output.Manager {
	emailConfig := output.EmailConfig{
		SMTPHost:       config.Email.SMTPHost,
		SMTPPort:       config.Email.SMTPPort,
		FromEmail:      config.Email.FromEmail,
		FromName:       config.Email.FromName,
		PasswordEnvVar: config.Email.PasswordEnvVar,
		SubjectPrefix:  config.Email.SubjectPrefix,
	}
	return output.NewManager(
		output.NewFileOutput("responses"),
		output.NewEmailOutputWithRecipients(emailConfig, emailTo),
	)
}

// OptimizePrompt applies master prompt optimization using the given provider
func OptimizePrompt(request string, provider Provider, sessionID string) (string, error) {
	fmt.Printf("Creating final prompt for request %s\n", sessionID)

	developerInstructions := masterPrompt
	requestToPromptText := fmt.Sprintf("Create the best prompt to address the following request: %s", request)

	return provider.Send(requestToPromptText, &developerInstructions)
}

// NewSession generates a new session UUID
func NewSession() string {
	return uuid.NewString()
}

// runProvidersConcurrently sends the prompt to all providers concurrently and handles responses
func runProvidersConcurrently(providers []Provider, prompt, sessionID string, outputManager *output.Manager) {
	responses := make(chan Response)
	done := make(chan struct{})
	waitGroup := sync.WaitGroup{}

	for _, provider := range providers {
		waitGroup.Add(1)
		go func(p Provider) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "Panic from %s: %v\n%s\n", p.Name(), r, debug.Stack())
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
			outputManager.Send(response.Text, sessionID, response.Provider, response.ProviderType)
		}
		done <- struct{}{}
	}()

	waitGroup.Wait()
	close(responses)
	<-done
}
