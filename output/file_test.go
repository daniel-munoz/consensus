package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileOutput_Name(t *testing.T) {
	fileOutput := NewFileOutput("/tmp")
	if fileOutput.Name() != "file" {
		t.Errorf("Expected name 'file', got '%s'", fileOutput.Name())
	}
}

func TestFileOutput_Send(t *testing.T) {
	tempDir := t.TempDir()
	fileOutput := NewFileOutput(tempDir)

	content := "test content"
	sessionID := "session123"
	context := "openai"

	err := fileOutput.Send(content, sessionID, context)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedFilename := filepath.Join(tempDir, "id-session123-openai.txt")
	if _, err := os.Stat(expectedFilename); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist", expectedFilename)
	}

	fileContent, err := os.ReadFile(expectedFilename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(fileContent) != content {
		t.Errorf("Expected file content '%s', got '%s'", content, string(fileContent))
	}
}

func TestFileOutput_Send_InvalidDirectory(t *testing.T) {
	fileOutput := NewFileOutput("/invalid/directory/that/does/not/exist")

	err := fileOutput.Send("content", "session", "context")
	if err == nil {
		t.Error("Expected error for invalid directory, got nil")
	}
}

func TestFileOutput_Send_EmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	fileOutput := NewFileOutput(tempDir)

	err := fileOutput.Send("", "session123", "context")
	if err != nil {
		t.Fatalf("Expected no error for empty content, got %v", err)
	}

	expectedFilename := filepath.Join(tempDir, "id-session123-context.txt")
	fileContent, err := os.ReadFile(expectedFilename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if len(fileContent) != 0 {
		t.Errorf("Expected empty file, got content: '%s'", string(fileContent))
	}
}

func TestFileOutput_Send_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	fileOutput := NewFileOutput(tempDir)

	content := "Special chars: 你好 🌟 ñoël"
	sessionID := "session-with-dashes"
	context := "test_context"

	err := fileOutput.Send(content, sessionID, context)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedFilename := filepath.Join(tempDir, "id-session-with-dashes-test_context.txt")
	fileContent, err := os.ReadFile(expectedFilename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(fileContent) != content {
		t.Errorf("Expected file content '%s', got '%s'", content, string(fileContent))
	}
}

func TestFileOutput_Send_LongContent(t *testing.T) {
	tempDir := t.TempDir()
	fileOutput := NewFileOutput(tempDir)

	content := strings.Repeat("A", 10000) + "\n" + strings.Repeat("B", 10000)
	sessionID := "long-session"
	context := "long-context"

	err := fileOutput.Send(content, sessionID, context)
	if err != nil {
		t.Fatalf("Expected no error for long content, got %v", err)
	}

	expectedFilename := filepath.Join(tempDir, "id-long-session-long-context.txt")
	fileContent, err := os.ReadFile(expectedFilename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(fileContent) != content {
		t.Error("File content doesn't match expected long content")
	}
}
