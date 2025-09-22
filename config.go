package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	MQTT struct {
		Broker         string `yaml:"broker"`
		Topic          string `yaml:"topic"`
		Username       string `yaml:"username,omitempty"`
		Password       string `yaml:"password,omitempty"`
		ConnectTimeout string `yaml:"connect_timeout,omitempty"`
		PingTimeout    string `yaml:"ping_timeout,omitempty"`
	} `yaml:"mqtt"`
	Ntfy struct {
		URL        string `yaml:"url"`
		AuthToken  string `yaml:"auth_token,omitempty"`
		Priority   string `yaml:"priority,omitempty"`
		Timeout    string `yaml:"timeout,omitempty"`
		MaxRetries int    `yaml:"max_retries,omitempty"`
		RetryDelay string `yaml:"retry_delay,omitempty"`
	} `yaml:"ntfy"`
}

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	setDefaults(&config)

	// Normalize the broker URL
	normalizedBroker, err := normalizeBrokerURL(config.MQTT.Broker)
	if err != nil {
		return nil, fmt.Errorf("invalid MQTT broker URL: %w", err)
	}
	config.MQTT.Broker = normalizedBroker

	return &config, nil
}

// validateConfig checks that required fields are present
func validateConfig(config *Config) error {
	if config.MQTT.Broker == "" {
		return fmt.Errorf("mqtt.broker is required in config")
	}
	if config.MQTT.Topic == "" {
		return fmt.Errorf("mqtt.topic is required in config")
	}
	if config.Ntfy.URL == "" {
		return fmt.Errorf("ntfy.url is required in config")
	}
	return nil
}

// setDefaults sets default values for optional configuration fields
func setDefaults(config *Config) {
	if config.MQTT.ConnectTimeout == "" {
		config.MQTT.ConnectTimeout = "30s"
	}
	if config.MQTT.PingTimeout == "" {
		config.MQTT.PingTimeout = "10s"
	}
	if config.Ntfy.Timeout == "" {
		config.Ntfy.Timeout = "10s"
	}
	if config.Ntfy.MaxRetries == 0 {
		config.Ntfy.MaxRetries = 3
	}
	if config.Ntfy.RetryDelay == "" {
		config.Ntfy.RetryDelay = "1s"
	}
}

// GetMQTTConnectTimeout parses the MQTT connect timeout duration
func (c *Config) GetMQTTConnectTimeout() time.Duration {
	duration, err := time.ParseDuration(c.MQTT.ConnectTimeout)
	if err != nil {
		return 30 * time.Second // fallback default
	}
	return duration
}

// GetMQTTPingTimeout parses the MQTT ping timeout duration
func (c *Config) GetMQTTPingTimeout() time.Duration {
	duration, err := time.ParseDuration(c.MQTT.PingTimeout)
	if err != nil {
		return 10 * time.Second // fallback default
	}
	return duration
}

// GetNtfyTimeout parses the Ntfy timeout duration
func (c *Config) GetNtfyTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Ntfy.Timeout)
	if err != nil {
		return 10 * time.Second // fallback default
	}
	return duration
}

// GetNtfyRetryDelay parses the Ntfy retry delay duration
func (c *Config) GetNtfyRetryDelay() time.Duration {
	duration, err := time.ParseDuration(c.Ntfy.RetryDelay)
	if err != nil {
		return 1 * time.Second // fallback default
	}
	return duration
}

// normalizeBrokerURL adds default protocol (tcp://) and port (1883) if not specified
func normalizeBrokerURL(broker string) (string, error) {
	if broker == "" {
		return "", fmt.Errorf("broker URL cannot be empty")
	}

	originalBroker := broker

	// If no protocol is specified, add tcp://
	if !strings.Contains(broker, "://") {
		broker = "tcp://" + broker
	}

	// Parse the URL to validate and potentially add default port
	parsedURL, err := url.Parse(broker)
	if err != nil {
		return "", fmt.Errorf("failed to parse broker URL '%s': %w", originalBroker, err)
	}

	// Only add default port for tcp and ssl schemes
	if (parsedURL.Scheme == "tcp" || parsedURL.Scheme == "ssl") && parsedURL.Port() == "" {
		// Check if we have just a hostname or hostname:port
		host := parsedURL.Host
		if !strings.Contains(host, ":") {
			// No port specified, add default MQTT port
			parsedURL.Host = host + ":1883"
		}
	}

	return parsedURL.String(), nil
}

// FlagConfig holds command-line flag values
type FlagConfig struct {
	MQTTBroker     string
	MQTTTopic      string
	MQTTUsername   string
	MQTTPassword   string
	NtfyURL        string
	NtfyAuthToken  string
	NtfyPriority   string
}

// LoadConfigWithOverrides loads configuration with flag/env variable overrides
func LoadConfigWithOverrides(configPath string, flags FlagConfig) (*Config, error) {
	var config Config

	// Load config file if provided
	if configPath != "" {
		fileConfig, err := LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		config = *fileConfig
	} else {
		// Set defaults when no config file
		setDefaults(&config)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	// Apply flag overrides
	applyFlagOverrides(&config, flags)

	// Validate required fields are present
	if err := validateRequiredConfig(&config); err != nil {
		return nil, err
	}

	// Normalize broker URL
	if config.MQTT.Broker != "" {
		normalizedBroker, err := normalizeBrokerURL(config.MQTT.Broker)
		if err != nil {
			return nil, fmt.Errorf("invalid MQTT broker URL: %w", err)
		}
		config.MQTT.Broker = normalizedBroker
	}

	return &config, nil
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(config *Config) {
	if val := os.Getenv("MQTT_USERNAME"); val != "" {
		config.MQTT.Username = val
	}
	if val := os.Getenv("MQTT_PASSWORD"); val != "" {
		config.MQTT.Password = val
	}
	if val := os.Getenv("NTFY_AUTH_TOKEN"); val != "" {
		config.Ntfy.AuthToken = val
	}
}

// applyFlagOverrides applies command-line flag overrides
func applyFlagOverrides(config *Config, flags FlagConfig) {
	if flags.MQTTBroker != "" {
		config.MQTT.Broker = flags.MQTTBroker
	}
	if flags.MQTTTopic != "" {
		config.MQTT.Topic = flags.MQTTTopic
	}
	if flags.MQTTUsername != "" {
		config.MQTT.Username = flags.MQTTUsername
	}
	if flags.MQTTPassword != "" {
		config.MQTT.Password = flags.MQTTPassword
	}
	if flags.NtfyURL != "" {
		config.Ntfy.URL = flags.NtfyURL
	}
	if flags.NtfyAuthToken != "" {
		config.Ntfy.AuthToken = flags.NtfyAuthToken
	}
	if flags.NtfyPriority != "" {
		config.Ntfy.Priority = flags.NtfyPriority
	}
}

// validateRequiredConfig validates that all required configuration is present
func validateRequiredConfig(config *Config) error {
	if config.MQTT.Broker == "" {
		return fmt.Errorf("MQTT broker is required (use --mqtt-broker flag, config file, or both)")
	}
	if config.MQTT.Topic == "" {
		return fmt.Errorf("MQTT topic is required (use --mqtt-topic flag, config file, or both)")
	}
	if config.Ntfy.URL == "" {
		return fmt.Errorf("Ntfy URL is required (use --ntfy-url flag, config file, or both)")
	}
	return nil
}
