package output

import (
	"fmt"
	"os"
)

// dirPerms are the permissions used when creating the output directory
const dirPerms = 0755

type FileOutput struct {
	BaseDir          string
	IgnoredProducers map[string]struct{}
}

func NewFileOutput(baseDir string) *FileOutput {
	return &FileOutput{
		BaseDir:          baseDir,
		IgnoredProducers: make(map[string]struct{}),
	}
}

func (f *FileOutput) Name() string {
	return "file"
}

func (f *FileOutput) WithIgnored(producers ...string) Writer {
	if f.IgnoredProducers == nil {
		f.IgnoredProducers = make(map[string]struct{})
	}
	for _, producer := range producers {
		f.IgnoredProducers[producer] = struct{}{}
	}
	return f
}

func (f *FileOutput) Send(content, sessionID, producerName, _ string) error {
	if _, ignored := f.IgnoredProducers[producerName]; ignored {
		return nil
	}
	if err := os.MkdirAll(f.BaseDir, dirPerms); err != nil {
		return err
	}
	filename := fmt.Sprintf("%s/id-%s-%s.txt", f.BaseDir, sessionID, producerName)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}
