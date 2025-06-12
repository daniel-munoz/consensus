package output

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	mock1 := NewMockWriter("writer1")
	mock2 := NewMockWriter("writer2")

	manager := NewManager(mock1, mock2)

	if len(manager.outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(manager.outputs))
	}
}

func TestNewManager_NoOutputs(t *testing.T) {
	manager := NewManager()

	if len(manager.outputs) != 0 {
		t.Errorf("Expected 0 outputs, got %d", len(manager.outputs))
	}
}

func TestManager_AddOutput(t *testing.T) {
	manager := NewManager()
	mock := NewMockWriter("test")

	manager.AddOutput(mock)

	if len(manager.outputs) != 1 {
		t.Errorf("Expected 1 output after add, got %d", len(manager.outputs))
	}

	if manager.outputs[0] != mock {
		t.Error("Added output doesn't match expected mock")
	}
}

func TestManager_Send(t *testing.T) {
	mock1 := NewMockWriter("writer1")
	mock2 := NewMockWriter("writer2")
	mock3 := NewMockWriter("writer3")

	manager := NewManager(mock1, mock2, mock3)

	content := "test content"
	sessionID := "session123"
	producer := "test-producer"

	manager.Send(content, sessionID, producer, producer)

	writers := []*MockWriter{mock1, mock2, mock3}
	for i, writer := range writers {
		if writer.callCount != 1 {
			t.Errorf("Writer %d: expected call count 1, got %d", i, writer.callCount)
		}
		if writer.lastContent != content {
			t.Errorf("Writer %d: expected content '%s', got '%s'", i, content, writer.lastContent)
		}
		if writer.lastSession != sessionID {
			t.Errorf("Writer %d: expected session '%s', got '%s'", i, sessionID, writer.lastSession)
		}
		if writer.lastProducer != producer {
			t.Errorf("Writer %d: expected producer '%s', got '%s'", i, producer, writer.lastProducer)
		}
	}
}

func TestManager_Send_WithErrors(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	mock1 := NewMockWriter("writer1")
	mock2 := NewMockWriter("writer2")
	mock3 := NewMockWriter("writer3")

	mock2.SetError(errors.New("test error"))

	manager := NewManager(mock1, mock2, mock3)
	manager.Send("content", "session", "producer", "producer")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Error sending to writer2 output: test error") {
		t.Error("Expected error message for writer2 not found in stderr")
	}

	if mock1.callCount != 1 {
		t.Errorf("Writer1: expected call count 1, got %d", mock1.callCount)
	}
	if mock2.callCount != 1 {
		t.Errorf("Writer2: expected call count 1, got %d", mock2.callCount)
	}
	if mock3.callCount != 1 {
		t.Errorf("Writer3: expected call count 1, got %d", mock3.callCount)
	}
}

func TestManager_Send_Concurrent(t *testing.T) {
	const numWriters = 10
	writers := make([]*MockWriter, numWriters)
	for i := 0; i < numWriters; i++ {
		writers[i] = NewMockWriter("writer" + string(rune('A'+i)))
	}

	var interfaceWriters []Writer
	for _, w := range writers {
		interfaceWriters = append(interfaceWriters, w)
	}

	manager := NewManager(interfaceWriters...)

	start := time.Now()
	manager.Send("test content", "session", "producer", "producer")
	duration := time.Since(start)

	for i, writer := range writers {
		if writer.callCount != 1 {
			t.Errorf("Writer %d: expected call count 1, got %d", i, writer.callCount)
		}
	}

	if duration > time.Second {
		t.Errorf("Send took too long: %v (expected concurrent execution)", duration)
	}
}

func TestManager_Send_RaceCondition(t *testing.T) {
	mock := NewMockWriter("concurrent-writer")
	manager := NewManager(mock)

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			manager.Send("content", "session", "producer", "producer")
		}(i)
	}

	wg.Wait()

	if mock.callCount != numGoroutines {
		t.Errorf("Expected call count %d, got %d", numGoroutines, mock.callCount)
	}
}

func TestManager_Send_EmptyManager(t *testing.T) {
	manager := NewManager()

	manager.Send("content", "session", "producer", "producer")
}
