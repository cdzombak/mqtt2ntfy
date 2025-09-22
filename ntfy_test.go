package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewNtfyClient(config, logger)
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	err := ForwardToNtfy(server.URL, "test message", "", "", config, logger)
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	err := ForwardToNtfy(server.URL, "test message", "", "", config, logger)
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	err := ForwardToNtfy("invalid-url", "test message", "", "", config, logger)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestIsWildcardTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		expected bool
	}{
		{
			name:     "wildcard topic",
			topic:    "my/notifications/#",
			expected: true,
		},
		{
			name:     "simple wildcard",
			topic:    "#",
			expected: true,
		},
		{
			name:     "root wildcard",
			topic:    "/#",
			expected: true,
		},
		{
			name:     "regular topic",
			topic:    "my/notifications/alerts",
			expected: false,
		},
		{
			name:     "topic ending with hash but not wildcard",
			topic:    "my/notifications#",
			expected: false,
		},
		{
			name:     "empty topic",
			topic:    "",
			expected: false,
		},
		{
			name:     "single character",
			topic:    "#",
			expected: true,
		},
		{
			name:     "only slash",
			topic:    "/",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWildcardTopic(tt.topic)
			if result != tt.expected {
				t.Errorf("IsWildcardTopic(%s) = %v, want %v", tt.topic, result, tt.expected)
			}
		})
	}
}

func TestExtractNtfyTopicFromMQTT(t *testing.T) {
	tests := []struct {
		name             string
		subscriptionTopic string
		receivedTopic    string
		expectedTopic    string
		expectError      bool
	}{
		{
			name:             "valid single level extraction",
			subscriptionTopic: "my/notifications/#",
			receivedTopic:    "my/notifications/alerts",
			expectedTopic:    "alerts",
			expectError:      false,
		},
		{
			name:             "valid single level with underscore",
			subscriptionTopic: "home/sensors/#",
			receivedTopic:    "home/sensors/temperature_sensor",
			expectedTopic:    "temperature_sensor",
			expectError:      false,
		},
		{
			name:             "valid root level wildcard",
			subscriptionTopic: "#",
			receivedTopic:    "alerts",
			expectedTopic:    "alerts",
			expectError:      false,
		},
		{
			name:             "non-wildcard subscription",
			subscriptionTopic: "my/notifications/alerts",
			receivedTopic:    "my/notifications/alerts",
			expectedTopic:    "",
			expectError:      true,
		},
		{
			name:             "mismatched base pattern",
			subscriptionTopic: "my/notifications/#",
			receivedTopic:    "other/notifications/alerts",
			expectedTopic:    "",
			expectError:      true,
		},
		{
			name:             "no additional level",
			subscriptionTopic: "my/notifications/#",
			receivedTopic:    "my/notifications",
			expectedTopic:    "",
			expectError:      true,
		},
		{
			name:             "multiple levels beyond wildcard",
			subscriptionTopic: "my/notifications/#",
			receivedTopic:    "my/notifications/alerts/critical",
			expectedTopic:    "",
			expectError:      true,
		},
		{
			name:             "exact match with trailing slash",
			subscriptionTopic: "my/notifications/#",
			receivedTopic:    "my/notifications/",
			expectedTopic:    "",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractNtfyTopicFromMQTT(tt.subscriptionTopic, tt.receivedTopic)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractNtfyTopicFromMQTT() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ExtractNtfyTopicFromMQTT() unexpected error: %v", err)
				}
				if result != tt.expectedTopic {
					t.Errorf("ExtractNtfyTopicFromMQTT() = %s, want %s", result, tt.expectedTopic)
				}
			}
		})
	}
}

func TestBuildNtfyURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		ntfyTopic   string
		expectedURL string
		expectError bool
	}{
		{
			name:        "valid URL construction",
			baseURL:     "https://ntfy.sh",
			ntfyTopic:   "alerts",
			expectedURL: "https://ntfy.sh/alerts",
			expectError: false,
		},
		{
			name:        "base URL with trailing slash",
			baseURL:     "https://ntfy.sh/",
			ntfyTopic:   "alerts",
			expectedURL: "https://ntfy.sh/alerts",
			expectError: false,
		},
		{
			name:        "custom domain",
			baseURL:     "https://notifications.example.com",
			ntfyTopic:   "server_alerts",
			expectedURL: "https://notifications.example.com/server_alerts",
			expectError: false,
		},
		{
			name:        "empty base URL",
			baseURL:     "",
			ntfyTopic:   "alerts",
			expectedURL: "",
			expectError: true,
		},
		{
			name:        "empty ntfy topic",
			baseURL:     "https://ntfy.sh",
			ntfyTopic:   "",
			expectedURL: "",
			expectError: true,
		},
		{
			name:        "topic with special characters",
			baseURL:     "https://ntfy.sh",
			ntfyTopic:   "server-alerts_2024",
			expectedURL: "https://ntfy.sh/server-alerts_2024",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildNtfyURL(tt.baseURL, tt.ntfyTopic)

			if tt.expectError {
				if err == nil {
					t.Errorf("BuildNtfyURL() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("BuildNtfyURL() unexpected error: %v", err)
				}
				if result != tt.expectedURL {
					t.Errorf("BuildNtfyURL() = %s, want %s", result, tt.expectedURL)
				}
			}
		})
	}
}

func TestParseMessagePriority(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		defaultPriority string
		expectedMessage string
		expectedPriority string
	}{
		{
			name:            "numeric priority 1",
			message:         "1|Quiet notification",
			defaultPriority: "3",
			expectedMessage: "Quiet notification",
			expectedPriority: "1",
		},
		{
			name:            "numeric priority 2",
			message:         "2|Low priority message",
			defaultPriority: "3",
			expectedMessage: "Low priority message",
			expectedPriority: "2",
		},
		{
			name:            "numeric priority 3",
			message:         "3|Normal priority",
			defaultPriority: "1",
			expectedMessage: "Normal priority",
			expectedPriority: "3",
		},
		{
			name:            "numeric priority 4",
			message:         "4|High priority",
			defaultPriority: "3",
			expectedMessage: "High priority",
			expectedPriority: "4",
		},
		{
			name:            "numeric priority 5",
			message:         "5|Critical urgent alert",
			defaultPriority: "3",
			expectedMessage: "Critical urgent alert",
			expectedPriority: "5",
		},
		{
			name:            "default priority g",
			message:         "g|Default priority message",
			defaultPriority: "2",
			expectedMessage: "Default priority message",
			expectedPriority: "2",
		},
		{
			name:            "high priority y",
			message:         "y|Yellow alert",
			defaultPriority: "3",
			expectedMessage: "Yellow alert",
			expectedPriority: "4",
		},
		{
			name:            "high priority o",
			message:         "o|Orange alert",
			defaultPriority: "3",
			expectedMessage: "Orange alert",
			expectedPriority: "4",
		},
		{
			name:            "urgent priority r",
			message:         "r|Red alert! Emergency!",
			defaultPriority: "3",
			expectedMessage: "Red alert! Emergency!",
			expectedPriority: "5",
		},
		{
			name:            "no prefix",
			message:         "Regular message without prefix",
			defaultPriority: "3",
			expectedMessage: "Regular message without prefix",
			expectedPriority: "3",
		},
		{
			name:            "invalid prefix",
			message:         "x|Invalid prefix",
			defaultPriority: "3",
			expectedMessage: "x|Invalid prefix",
			expectedPriority: "3",
		},
		{
			name:            "message too short",
			message:         "1",
			defaultPriority: "3",
			expectedMessage: "1",
			expectedPriority: "3",
		},
		{
			name:            "no pipe separator",
			message:         "1-No pipe separator",
			defaultPriority: "3",
			expectedMessage: "1-No pipe separator",
			expectedPriority: "3",
		},
		{
			name:            "empty message after prefix",
			message:         "5|",
			defaultPriority: "3",
			expectedMessage: "",
			expectedPriority: "5",
		},
		{
			name:            "message with multiple pipes",
			message:         "2|Message with | multiple | pipes",
			defaultPriority: "3",
			expectedMessage: "Message with | multiple | pipes",
			expectedPriority: "2",
		},
		{
			name:            "empty default priority",
			message:         "g|Message with empty default",
			defaultPriority: "",
			expectedMessage: "Message with empty default",
			expectedPriority: "",
		},
		{
			name:            "numeric priority with special characters",
			message:         "4|Message with émojis 🚨 and ünícodé",
			defaultPriority: "3",
			expectedMessage: "Message with émojis 🚨 and ünícodé",
			expectedPriority: "4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanedMessage, priority := ParseMessagePriority(tt.message, tt.defaultPriority)

			if cleanedMessage != tt.expectedMessage {
				t.Errorf("ParseMessagePriority() cleanedMessage = %s, want %s", cleanedMessage, tt.expectedMessage)
			}
			if priority != tt.expectedPriority {
				t.Errorf("ParseMessagePriority() priority = %s, want %s", priority, tt.expectedPriority)
			}
		})
	}
}
