package main

import (
	"fmt"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
mqtt:
  broker: "tcp://localhost:1883"
  topic: "test/topic"
  connect_timeout: "30s"
  ping_timeout: "10s"
ntfy:
  url: "https://ntfy.sh/test"
  auth_token: "secret"
  timeout: "15s"
  max_retries: 5
  retry_delay: "2s"
`

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.MQTT.Broker != "tcp://localhost:1883" {
		t.Errorf("Expected MQTT.Broker 'tcp://localhost:1883', got '%s'", config.MQTT.Broker)
	}
	if config.MQTT.Topic != "test/topic" {
		t.Errorf("Expected MQTT.Topic 'test/topic', got '%s'", config.MQTT.Topic)
	}
	if config.Ntfy.URL != "https://ntfy.sh/test" {
		t.Errorf("Expected Ntfy.URL 'https://ntfy.sh/test', got '%s'", config.Ntfy.URL)
	}
	if config.Ntfy.AuthToken != "secret" {
		t.Errorf("Expected Ntfy.AuthToken 'secret', got '%s'", config.Ntfy.AuthToken)
	}
	if config.MQTT.ConnectTimeout != "30s" {
		t.Errorf("Expected MQTT.ConnectTimeout '30s', got '%s'", config.MQTT.ConnectTimeout)
	}
	if config.MQTT.PingTimeout != "10s" {
		t.Errorf("Expected MQTT.PingTimeout '10s', got '%s'", config.MQTT.PingTimeout)
	}
	if config.Ntfy.Timeout != "15s" {
		t.Errorf("Expected Ntfy.Timeout '15s', got '%s'", config.Ntfy.Timeout)
	}
	if config.Ntfy.MaxRetries != 5 {
		t.Errorf("Expected Ntfy.MaxRetries 5, got %d", config.Ntfy.MaxRetries)
	}
	if config.Ntfy.RetryDelay != "2s" {
		t.Errorf("Expected Ntfy.RetryDelay '2s', got '%s'", config.Ntfy.RetryDelay)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	invalidYAML := `invalid: yaml: content: [`
	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   error
	}{
		{
			name: "valid config",
			config: Config{},
			want:   nil,
		},
		{
			name: "missing mqtt.broker",
			config: Config{
				MQTT: struct {
					Broker         string `yaml:"broker"`
					Topic          string `yaml:"topic"`
					Username       string `yaml:"username,omitempty"`
					Password       string `yaml:"password,omitempty"`
					ConnectTimeout string `yaml:"connect_timeout,omitempty"`
					PingTimeout    string `yaml:"ping_timeout,omitempty"`
				}{Topic: "test/topic"},
				Ntfy: struct {
					URL        string `yaml:"url"`
					AuthToken  string `yaml:"auth_token,omitempty"`
					Priority   string `yaml:"priority,omitempty"`
					Timeout    string `yaml:"timeout,omitempty"`
					MaxRetries int    `yaml:"max_retries,omitempty"`
					RetryDelay string `yaml:"retry_delay,omitempty"`
				}{URL: "https://ntfy.sh/test"},
			},
			want: fmt.Errorf("mqtt.broker is required in config"),
		},
		{
			name: "missing mqtt.topic",
			config: Config{
				MQTT: struct {
					Broker         string `yaml:"broker"`
					Topic          string `yaml:"topic"`
					Username       string `yaml:"username,omitempty"`
					Password       string `yaml:"password,omitempty"`
					ConnectTimeout string `yaml:"connect_timeout,omitempty"`
					PingTimeout    string `yaml:"ping_timeout,omitempty"`
				}{Broker: "tcp://localhost:1883"},
				Ntfy: struct {
					URL        string `yaml:"url"`
					AuthToken  string `yaml:"auth_token,omitempty"`
					Priority   string `yaml:"priority,omitempty"`
					Timeout    string `yaml:"timeout,omitempty"`
					MaxRetries int    `yaml:"max_retries,omitempty"`
					RetryDelay string `yaml:"retry_delay,omitempty"`
				}{URL: "https://ntfy.sh/test"},
			},
			want: fmt.Errorf("mqtt.topic is required in config"),
		},
		{
			name: "missing ntfy.url",
			config: Config{
				MQTT: struct {
					Broker         string `yaml:"broker"`
					Topic          string `yaml:"topic"`
					Username       string `yaml:"username,omitempty"`
					Password       string `yaml:"password,omitempty"`
					ConnectTimeout string `yaml:"connect_timeout,omitempty"`
					PingTimeout    string `yaml:"ping_timeout,omitempty"`
				}{Broker: "tcp://localhost:1883", Topic: "test/topic"},
				Ntfy: struct {
					URL        string `yaml:"url"`
					AuthToken  string `yaml:"auth_token,omitempty"`
					Priority   string `yaml:"priority,omitempty"`
					Timeout    string `yaml:"timeout,omitempty"`
					MaxRetries int    `yaml:"max_retries,omitempty"`
					RetryDelay string `yaml:"retry_delay,omitempty"`
				}{},
			},
			want: fmt.Errorf("ntfy.url is required in config"),
		},
	}

	// Set up valid config for first test
	tests[0].config.MQTT.Broker = "tcp://localhost:1883"
	tests[0].config.MQTT.Topic = "test/topic"
	tests[0].config.Ntfy.URL = "https://ntfy.sh/test"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.want == nil && err != nil {
				t.Errorf("validateConfig() error = %v, want nil", err)
			}
			if tt.want != nil && err == nil {
				t.Errorf("validateConfig() error = nil, want %v", tt.want)
			}
			if tt.want != nil && err != nil && err.Error() != tt.want.Error() {
				t.Errorf("validateConfig() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNormalizeBrokerURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "hostname only",
			input:    "localhost",
			expected: "tcp://localhost:1883",
			wantErr:  false,
		},
		{
			name:     "hostname with port",
			input:    "localhost:1884",
			expected: "tcp://localhost:1884",
			wantErr:  false,
		},
		{
			name:     "full tcp URL",
			input:    "tcp://localhost:1883",
			expected: "tcp://localhost:1883",
			wantErr:  false,
		},
		{
			name:     "tcp without port",
			input:    "tcp://localhost",
			expected: "tcp://localhost:1883",
			wantErr:  false,
		},
		{
			name:     "ssl without port",
			input:    "ssl://localhost",
			expected: "ssl://localhost:1883",
			wantErr:  false,
		},
		{
			name:     "ssl with port",
			input:    "ssl://localhost:8883",
			expected: "ssl://localhost:8883",
			wantErr:  false,
		},
		{
			name:     "websocket URL unchanged",
			input:    "ws://localhost:8080/mqtt",
			expected: "ws://localhost:8080/mqtt",
			wantErr:  false,
		},
		{
			name:     "secure websocket URL unchanged",
			input:    "wss://localhost:8080/mqtt",
			expected: "wss://localhost:8080/mqtt",
			wantErr:  false,
		},
		{
			name:     "IP address",
			input:    "192.168.1.100",
			expected: "tcp://192.168.1.100:1883",
			wantErr:  false,
		},
		{
			name:     "IP address with port",
			input:    "192.168.1.100:1884",
			expected: "tcp://192.168.1.100:1884",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "FQDN hostname",
			input:    "mqtt.example.com",
			expected: "tcp://mqtt.example.com:1883",
			wantErr:  false,
		},
		{
			name:     "FQDN with port",
			input:    "mqtt.example.com:1884",
			expected: "tcp://mqtt.example.com:1884",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeBrokerURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("normalizeBrokerURL() expected error for input '%s', got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("normalizeBrokerURL() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if result != tt.expected {
					t.Errorf("normalizeBrokerURL() = '%s', want '%s'", result, tt.expected)
				}
			}
		})
	}
}

func TestLoadConfigWithOverrides(t *testing.T) {
	tests := []struct {
		name        string
		flags       FlagConfig
		expectError bool
	}{
		{
			name: "flags only (no config file)",
			flags: FlagConfig{
				MQTTBroker: "flag-broker",
				MQTTTopic:  "flag-topic",
				NtfyURL:    "https://flag.example.com",
			},
			expectError: false,
		},
		{
			name:        "missing required config",
			flags:       FlagConfig{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadConfigWithOverrides("", tt.flags)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Errorf("Config should not be nil")
			}
		})
	}
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	config := Config{}
	setDefaults(&config)
	config.MQTT.Username = "base-user"
	config.MQTT.Password = "base-pass"
	config.Ntfy.AuthToken = "base-token"

	// Set environment variables
	os.Setenv("MQTT_USERNAME", "env-user")
	os.Setenv("MQTT_PASSWORD", "env-pass")
	os.Setenv("NTFY_AUTH_TOKEN", "env-token")
	defer func() {
		os.Unsetenv("MQTT_USERNAME")
		os.Unsetenv("MQTT_PASSWORD")
		os.Unsetenv("NTFY_AUTH_TOKEN")
	}()

	applyEnvOverrides(&config)

	if config.MQTT.Username != "env-user" {
		t.Errorf("MQTT.Username = %s, want env-user", config.MQTT.Username)
	}
	if config.MQTT.Password != "env-pass" {
		t.Errorf("MQTT.Password = %s, want env-pass", config.MQTT.Password)
	}
	if config.Ntfy.AuthToken != "env-token" {
		t.Errorf("Ntfy.AuthToken = %s, want env-token", config.Ntfy.AuthToken)
	}
}
