package file

import (
	"os"
)

// CreateFile creates a file named after the provider (e.g., "OpenAI.txt")
// and writes the given text content to it
func Create(provider string, content string) error {
	filename := provider + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}
