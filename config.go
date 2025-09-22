package main

import (
	"fmt"
	"os"
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
