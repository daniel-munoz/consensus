package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/daniel-munoz/consensus/ai"
	"github.com/daniel-munoz/consensus/file"
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
	request := readPrompt()

	var (
		prompt    string
		responses = make(chan Response)
		done      = make(chan struct{})
		err       error
	)

	developerInstructions := masterPrompt

	fmt.Println("Creating final prompt")

	requestToPromptText := fmt.Sprintf("Create the best prompt to address the following request: %s", request)
	prompt, err = ai.OpenAI{}.Send(requestToPromptText, &developerInstructions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(prompt)

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
			if err := file.Create(response.Provider, response.Text); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file for %s: %v\n", response.Provider, err)
			}
		}
		done <- struct{}{}
	}()

	waitGroup.Wait()
	close(responses)
	<-done
}
