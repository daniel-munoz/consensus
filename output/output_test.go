package output

import (
	"errors"
	"sync"
	"testing"
)

type MockWriter struct {
	name         string
	sendError    error
	mu           sync.Mutex
	callCount    int
	lastContent  string
	lastSession  string
	lastProducer string
	ignored      map[string]struct{}
}

func NewMockWriter(name string) *MockWriter {
	return &MockWriter{name: name}
}

func (m *MockWriter) Name() string {
	return m.name
}

func (m *MockWriter) WithIgnored(producers ...string) Writer {
	if m.ignored == nil {
		m.ignored = make(map[string]struct{})
	}
	for _, p := range producers {
		m.ignored[p] = struct{}{}
	}
	return m
}

func (m *MockWriter) Send(content, sessionID, producer, producerType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	if _, ignored := m.ignored[producer]; ignored {
		return nil
	}
	m.lastContent = content
	m.lastSession = sessionID
	m.lastProducer = producer
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

	err := mock.Send("test content", "session123", "producer456", "producer456")
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

	if mock.lastProducer != "producer456" {
		t.Errorf("Expected producer 'producer456', got '%s'", mock.lastProducer)
	}

	testError := errors.New("test error")
	mock.SetError(testError)
	err = mock.Send("", "", "", "")
	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}
}

func TestWithIgnored_SingleProducer(t *testing.T) {
	mock := NewMockWriter("test-writer")

	writer := mock.WithIgnored("ignored-producer")
	if writer != mock {
		t.Error("WithIgnored should return the same writer instance")
	}

	err := mock.Send("content", "session", "ignored-producer", "ignored-producer")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mock.callCount != 1 {
		t.Errorf("Expected call count 1, got %d", mock.callCount)
	}

	if mock.lastContent != "" {
		t.Errorf("Expected empty content for ignored producer, got '%s'", mock.lastContent)
	}

	err = mock.Send("content", "session", "active-producer", "active-producer")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mock.callCount != 2 {
		t.Errorf("Expected call count 2, got %d", mock.callCount)
	}

	if mock.lastContent != "content" {
		t.Errorf("Expected content 'content', got '%s'", mock.lastContent)
	}

	if mock.lastProducer != "active-producer" {
		t.Errorf("Expected producer 'active-producer', got '%s'", mock.lastProducer)
	}
}

func TestWithIgnored_MultipleProducers(t *testing.T) {
	mock := NewMockWriter("test-writer")

	mock.WithIgnored("producer1", "producer2", "producer3")

	testCases := []struct {
		producer     string
		shouldIgnore bool
	}{
		{"producer1", true},
		{"producer2", true},
		{"producer3", true},
		{"producer4", false},
		{"active-producer", false},
	}

	for _, tc := range testCases {
		initialCount := mock.callCount
		initialContent := mock.lastContent

		err := mock.Send("test content", "session", tc.producer, tc.producer)
		if err != nil {
			t.Errorf("Expected no error for producer %s, got %v", tc.producer, err)
		}

		if tc.shouldIgnore {
			if mock.lastContent != initialContent {
				t.Errorf("Producer %s should be ignored, but content was updated", tc.producer)
			}
		} else {
			if mock.lastContent != "test content" {
				t.Errorf("Producer %s should not be ignored, expected content 'test content', got '%s'", tc.producer, mock.lastContent)
			}
			if mock.lastProducer != tc.producer {
				t.Errorf("Producer %s should not be ignored, expected producer '%s', got '%s'", tc.producer, tc.producer, mock.lastProducer)
			}
		}

		if mock.callCount != initialCount+1 {
			t.Errorf("Call count should increment regardless of ignore status for producer %s", tc.producer)
		}
	}
}

func TestWithIgnored_ChainedCalls(t *testing.T) {
	mock := NewMockWriter("test-writer")

	writer := mock.WithIgnored("producer1").WithIgnored("producer2")
	if writer != mock {
		t.Error("Chained WithIgnored calls should return the same writer instance")
	}

	mock.Send("content", "session", "producer1", "producer1")
	if mock.lastContent != "" {
		t.Error("producer1 should be ignored")
	}

	mock.Send("content", "session", "producer2", "producer2")
	if mock.lastContent != "" {
		t.Error("producer2 should be ignored")
	}

	mock.Send("content", "session", "producer3", "producer3")
	if mock.lastContent != "content" {
		t.Errorf("producer3 should not be ignored, expected 'content', got '%s'", mock.lastContent)
	}
}
