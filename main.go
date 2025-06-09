package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/daniel-munoz/consensus/ai"
	"github.com/daniel-munoz/consensus/output"
	"github.com/google/uuid"
)

var inputReader io.Reader = os.Stdin

type Response struct {
	Text     string
	Provider string
}

type Provider interface {
	Name() string
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

func main() {
	promptFlag := flag.String("prompt", "", "Prompt text to use instead of interactive input")
	flag.StringVar(promptFlag, "p", "", "Prompt text to use instead of interactive input (shorthand)")
	flag.Parse()

	var (
		prompt    string
		request   string
		responses = make(chan Response)
		done      = make(chan struct{})
		err       error
	)

	if promptFlag != nil && *promptFlag != "" {
		request = *promptFlag
	} else {
		request = readPrompt()
	}

	// Generate UUID for this session
	id := uuid.NewString()

	// Initialize output manager with file and email output
	outputManager := output.NewManager(
		output.NewFileOutput("responses"),
		output.NewEmailOutput(),
	)

	outputManager.Send(request, id, "request")

	developerInstructions := masterPrompt

	fmt.Printf("Creating final prompt for request %s\n", id)

	requestToPromptText := fmt.Sprintf("Create the best prompt to address the following request: %s", request)
	prompt, err = ai.OpenAI{}.Send(requestToPromptText, &developerInstructions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputManager.Send(prompt, id, "prompt")

	waitGroup := sync.WaitGroup{}

	for _, provider := range []Provider{ai.OpenAI{}, ai.Anthropic{}, ai.Gemini{}} {
		waitGroup.Add(1)
		go func(p Provider) {
			defer waitGroup.Done()
			fmt.Printf("Consulting %s...\n", p.Name())
			response, err := p.Send(prompt, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from %s: %v\n", p.Name(), err)
				return
			}
			responses <- Response{Text: response, Provider: p.Name()}
			fmt.Printf("%s responded!\n", p.Name())
		}(provider)
	}

	go func() {
		for response := range responses {
			outputManager.Send(response.Text, id, response.Provider)
		}
		done <- struct{}{}
	}()

	waitGroup.Wait()
	close(responses)
	<-done
}
