package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
)

// NtfyClient interface for dependency injection and testing
type NtfyClient interface {
	SendMessage(url, message, authToken, priority string) error
}

// NtfyConfig holds configuration for the Ntfy client
type NtfyConfig struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// HTTPNtfyClient implements NtfyClient using HTTP
type HTTPNtfyClient struct {
	client *http.Client
	config NtfyConfig
	logger *slog.Logger
}

// NewNtfyClient creates a new HTTP Ntfy client with configuration
func NewNtfyClient(config NtfyConfig, logger *slog.Logger) *HTTPNtfyClient {
	return &HTTPNtfyClient{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
		logger: logger,
	}
}

// SendMessage implements NtfyClient interface with retry logic
func (n *HTTPNtfyClient) SendMessage(url, message, authToken, priority string) error {
	return retry.Do(
		func() error {
			return n.sendMessageOnce(url, message, authToken, priority)
		},
		retry.Attempts(uint(n.config.MaxRetries)),
		retry.Delay(n.config.RetryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(func(err error) bool {
			return isRetryableError(err)
		}),
	)
}

// sendMessageOnce performs a single HTTP request to send a message
func (n *HTTPNtfyClient) sendMessageOnce(url, message, authToken, priority string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(message))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	if priority != "" {
		req.Header.Set("Priority", priority)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the function
			n.logger.Warn("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("ntfy server returned status: %d (server error)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return retry.Unrecoverable(fmt.Errorf("ntfy server returned status: %d (client error)", resp.StatusCode))
	}

	return nil
}

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Retry on network errors
	if _, ok := err.(net.Error); ok {
		return true
	}

	// Retry on timeout errors
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	// Don't retry on unrecoverable errors
	if retry.IsRecoverable(err) {
		return false
	}

	return true
}

// ForwardToNtfy forwards a message to Ntfy with retry logic
func ForwardToNtfy(url, message, authToken, priority string, config NtfyConfig, logger *slog.Logger) error {
	client := NewNtfyClient(config, logger)
	return client.SendMessage(url, message, authToken, priority)
}

// IsWildcardTopic checks if the MQTT topic ends with /# or is just #
func IsWildcardTopic(mqttTopic string) bool {
	if mqttTopic == "#" {
		return true
	}
	return len(mqttTopic) >= 2 && mqttTopic[len(mqttTopic)-2:] == "/#"
}

// ExtractNtfyTopicFromMQTT extracts the Ntfy topic from an MQTT topic when using wildcards
// For example: "my/notifications/alerts" with subscription "my/notifications/#" returns "alerts"
func ExtractNtfyTopicFromMQTT(subscriptionTopic, receivedTopic string) (string, error) {
	if !IsWildcardTopic(subscriptionTopic) {
		return "", fmt.Errorf("subscription topic %s is not a wildcard topic", subscriptionTopic)
	}

	// Handle root wildcard case (subscription is just "#")
	if subscriptionTopic == "#" {
		// For root wildcard, the entire received topic is the ntfy topic
		// but it should only be one level (no slashes)
		if strings.Contains(receivedTopic, "/") {
			return "", fmt.Errorf("received topic %s has multiple levels beyond subscription pattern %s (one-level wildcard only)", receivedTopic, subscriptionTopic)
		}
		if receivedTopic == "" {
			return "", fmt.Errorf("received topic %s has no additional level beyond subscription pattern %s", receivedTopic, subscriptionTopic)
		}
		return receivedTopic, nil
	}

	// Remove the /# suffix to get the base pattern
	basePattern := subscriptionTopic[:len(subscriptionTopic)-2]

	// Check if the received topic starts with the base pattern
	if !strings.HasPrefix(receivedTopic, basePattern) {
		return "", fmt.Errorf("received topic %s does not match subscription pattern %s", receivedTopic, subscriptionTopic)
	}

	// Extract the remaining part after the base pattern
	remainder := receivedTopic[len(basePattern):]

	// Remove leading slash if present
	if len(remainder) > 0 && remainder[0] == '/' {
		remainder = remainder[1:]
	}

	// Check that there's exactly one more level (no additional slashes)
	if remainder == "" {
		return "", fmt.Errorf("received topic %s has no additional level beyond subscription pattern %s", receivedTopic, subscriptionTopic)
	}

	if strings.Contains(remainder, "/") {
		return "", fmt.Errorf("received topic %s has multiple levels beyond subscription pattern %s (one-level wildcard only)", receivedTopic, subscriptionTopic)
	}

	return remainder, nil
}

// BuildNtfyURL constructs the Ntfy URL using a base URL and extracted topic
// For example: base "https://ntfy.sh" + topic "alerts" = "https://ntfy.sh/alerts"
func BuildNtfyURL(baseURL, ntfyTopic string) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("base URL cannot be empty")
	}
	if ntfyTopic == "" {
		return "", fmt.Errorf("ntfy topic cannot be empty")
	}

	// Remove trailing slash from base URL if present
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return baseURL + "/" + ntfyTopic, nil
}
