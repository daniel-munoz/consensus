package output

import (
	"errors"
	"testing"
)

type MockWriter struct {
	name        string
	sendError   error
	callCount   int
	lastContent string
	lastSession string
	lastContext string
}

func NewMockWriter(name string) *MockWriter {
	return &MockWriter{name: name}
}

func (m *MockWriter) Name() string {
	return m.name
}

func (m *MockWriter) Send(content, sessionID, context string) error {
	m.callCount++
	m.lastContent = content
	m.lastSession = sessionID
	m.lastContext = context
	return m.sendError
}

func (m *MockWriter) SetError(err error) {
	m.sendError = err
}

func TestMockWriter(t *testing.T) {
	mock := NewMockWriter("test-writer")

	if mock.Name() != "test-writer" {
		t.Errorf("Expected name 'test-writer', got '%s'", mock.Name())
	}

	err := mock.Send("test content", "session123", "context456")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mock.callCount != 1 {
		t.Errorf("Expected call count 1, got %d", mock.callCount)
	}

	if mock.lastContent != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", mock.lastContent)
	}

	if mock.lastSession != "session123" {
		t.Errorf("Expected session 'session123', got '%s'", mock.lastSession)
	}

	if mock.lastContext != "context456" {
		t.Errorf("Expected context 'context456', got '%s'", mock.lastContext)
	}

	testError := errors.New("test error")
	mock.SetError(testError)
	err = mock.Send("", "", "")
	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}
}
