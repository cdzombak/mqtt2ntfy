package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"net/http"
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
