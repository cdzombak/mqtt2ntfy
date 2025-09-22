package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockNtfyClient for testing
type MockNtfyClient struct {
	sendError error
}

func (m *MockNtfyClient) SendMessage(url, message, authToken, priority string) error {
	return m.sendError
}

func TestNewNtfyClient(t *testing.T) {
	config := NtfyConfig{
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
	client := NewNtfyClient(config)
	if client == nil {
		t.Error("NewNtfyClient returned nil")
	}
}

func TestForwardToNtfySuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NtfyConfig{
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	err := ForwardToNtfy(server.URL, "test message", "", "", config)
	if err != nil {
		t.Errorf("ForwardToNtfy failed: %v", err)
	}
}

func TestForwardToNtfyFailure(t *testing.T) {
	// Create a test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := NtfyConfig{
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	err := ForwardToNtfy(server.URL, "test message", "", "", config)
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestForwardToNtfyInvalidURL(t *testing.T) {
	config := NtfyConfig{
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	err := ForwardToNtfy("invalid-url", "test message", "", "", config)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}
