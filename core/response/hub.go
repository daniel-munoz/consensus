package response

import (
	"errors"
)

// Hub is an interface that defines methods to interact with a collection of response providers.
type Hub interface {
	MasterPromptProviderName() string
	MasterPrompt() (string, error)
	Providers() []string
	ResponseFrom(string) (string, error)
}

// hubImpl is a concrete implementation of the Hub interface.
type hubImpl struct {
	promptProviderName string
	providerNames      []string
	promptResponse     DelayedResponse
	responses          map[string]DelayedResponse
}

// NewHub creates a new instance of Hub with the provided parameters.
func NewHub(promptProviderName string, providerNames []string, promptResponse DelayedResponse, responses map[string]DelayedResponse) Hub {
	return &hubImpl{
		promptProviderName: promptProviderName,
		providerNames:      providerNames,
		promptResponse:     promptResponse,
		responses:          responses,
	}
}

// MasterPromptProviderName returns the name of the master prompt provider.
func (h *hubImpl) MasterPromptProviderName() string {
	return h.promptProviderName
}

// MasterPrompt returns the master prompt response.
func (h *hubImpl) MasterPrompt() (string, error) {
	return h.promptResponse.Value()
}

// Providers returns a list of provider names.
func (h *hubImpl) Providers() []string {
	return h.providerNames
}

// ResponseFrom retrieves the response from a specific provider by name.
func (h *hubImpl) ResponseFrom(providerName string) (string, error) {
	if response, exists := h.responses[providerName]; exists {
		return response.Value()
	}
	return "", errors.New("provider not found: " + providerName)
}
