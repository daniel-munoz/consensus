package output

import (
	"fmt"
	"os"
)

type FileOutput struct {
	BaseDir string
}

func NewFileOutput(baseDir string) *FileOutput {
	return &FileOutput{BaseDir: baseDir}
}

func (f *FileOutput) Name() string {
	return "file"
}

func (f *FileOutput) Send(content, sessionID, context string) error {
	filename := fmt.Sprintf("%s/id-%s-%s.txt", f.BaseDir, sessionID, context)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

