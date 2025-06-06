package file

import (
	"fmt"
	"os"
)

// ResponseParams holds the parameters needed to create a file.
type ResponseParams struct {
	Folder  string
	ID      string
	Context string
	Content string
}

// CreateFile creates a file named after the provider with UUID (e.g., "OpenAI-uuid.txt")
// and writes the given text content to it
func Create(params ResponseParams) error {
	filename := fmt.Sprintf("%s/id-%s-%s.txt", params.Folder, params.ID, params.Context)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(params.Content)
	return err
}
