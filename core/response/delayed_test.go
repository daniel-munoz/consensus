package response

import (
	"errors"
	"testing"
	"time"
)

func TestDelayedResponse_Value_AlreadyResponded(t *testing.T) {
	dr := DelayedResponse{
		responded: true,
		value:     "test response",
		err:       nil,
	}

	value, err := dr.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value != "test response" {
		t.Errorf("Expected 'test response', got '%s'", value)
	}
}

func TestDelayedResponse_Value_AlreadyRespondedWithError(t *testing.T) {
	expectedErr := errors.New("test error")
	dr := DelayedResponse{
		responded: true,
		value:     "",
		err:       expectedErr,
	}

	value, err := dr.Value()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}
}

func TestDelayedResponse_Value_ReceiveFromValueQueue(t *testing.T) {
	valueQueue := make(chan string, 1)
	errorQueue := make(chan error, 1)
	
	dr := DelayedResponse{
		timeoutInMinutes: 1,
		valueQueue:       valueQueue,
		errorQueue:       errorQueue,
		responded:        false,
	}

	valueQueue <- "test response"

	value, err := dr.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value != "test response" {
		t.Errorf("Expected 'test response', got '%s'", value)
	}
}

func TestDelayedResponse_Value_ReceiveFromErrorQueue(t *testing.T) {
	valueQueue := make(chan string, 1)
	errorQueue := make(chan error, 1)
	expectedErr := errors.New("test error")
	
	dr := DelayedResponse{
		timeoutInMinutes: 1,
		valueQueue:       valueQueue,
		errorQueue:       errorQueue,
		responded:        false,
	}

	errorQueue <- expectedErr

	value, err := dr.Value()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}
}

func TestDelayedResponse_Value_Timeout(t *testing.T) {
	valueQueue := make(chan string)
	errorQueue := make(chan error)
	
	dr := DelayedResponse{
		timeoutInMinutes: 0, // Use 0 to make timeout immediate for testing
		valueQueue:       valueQueue,
		errorQueue:       errorQueue,
		responded:        false,
	}

	start := time.Now()
	value, err := dr.Value()
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if err.Error() != "timeout waiting for response" {
		t.Errorf("Expected 'timeout waiting for response', got '%s'", err.Error())
	}
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}
	if duration >= time.Minute {
		t.Error("Timeout took too long, expected immediate timeout with 0 minutes")
	}
}