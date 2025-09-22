package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cdzombak/heartbeat"
)

var version = "dev"

func main() {
	var configPath string
	var verbose bool
	var showVersion bool

	// MQTT flags
	var mqttBroker string
	var mqttTopic string
	var mqttUsername string
	var mqttPassword string

	// Ntfy flags
	var ntfyURL string
	var ntfyAuthToken string
	var ntfyPriority string

	flag.StringVar(&configPath, "config", "", "Path to YAML configuration file (optional if all required flags provided)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	// MQTT flags
	flag.StringVar(&mqttBroker, "mqtt-broker", "", "MQTT broker URL (e.g., localhost, tcp://localhost:1883)")
	flag.StringVar(&mqttTopic, "mqtt-topic", "", "MQTT topic to subscribe to")
	flag.StringVar(&mqttUsername, "mqtt-username", "", "MQTT username for authentication")
	flag.StringVar(&mqttPassword, "mqtt-password", "", "MQTT password for authentication")

	// Ntfy flags
	flag.StringVar(&ntfyURL, "ntfy-url", "", "Ntfy server URL (e.g., https://ntfy.sh/your-topic)")
	flag.StringVar(&ntfyAuthToken, "ntfy-token", "", "Ntfy authentication token")
	flag.StringVar(&ntfyPriority, "ntfy-priority", "", "Ntfy message priority (1-5)")

	flag.Parse()

	logger := SetupLogger(verbose)

	if showVersion {
		logger.Info("Version information", "version", version)
		os.Exit(0)
	}
	logger.Info("Starting mqtt2ntfy", "config", configPath)

	// Prepare flag configuration
	flags := FlagConfig{
		MQTTBroker:    mqttBroker,
		MQTTTopic:     mqttTopic,
		MQTTUsername:  mqttUsername,
		MQTTPassword:  mqttPassword,
		NtfyURL:       ntfyURL,
		NtfyAuthToken: ntfyAuthToken,
		NtfyPriority:  ntfyPriority,
	}

	config, err := LoadConfigWithOverrides(configPath, flags)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Config loaded successfully", "mqtt_broker", config.MQTT.Broker, "mqtt_topic", config.MQTT.Topic, "ntfy_url", config.Ntfy.URL)

	// Create Ntfy configuration
	ntfyConfig := NtfyConfig{
		Timeout:    config.GetNtfyTimeout(),
		MaxRetries: config.Ntfy.MaxRetries,
		RetryDelay: config.GetNtfyRetryDelay(),
	}

	// Connect to MQTT and subscribe
	mqttHandler, err := ConnectAndSubscribe(context.Background(), config.MQTT.Broker, config.MQTT.Topic, config.MQTT.Username, config.MQTT.Password, config.GetMQTTConnectTimeout(), config.GetMQTTPingTimeout(), func(topic string, payload []byte) {
		logger.Info("Received MQTT message", "topic", topic, "payload", string(payload))

		// Parse message for priority prefix and get cleaned message
		cleanedMessage, messagePriority := ParseMessagePriority(string(payload), config.Ntfy.Priority)
		if cleanedMessage != string(payload) {
			logger.Info("Extracted priority from message", "original", string(payload), "cleaned", cleanedMessage, "priority", messagePriority)
		}

		// Determine the Ntfy URL to use
		var ntfyURL string
		if IsWildcardTopic(config.MQTT.Topic) {
			// Extract Ntfy topic from MQTT topic for wildcard subscriptions
			ntfyTopic, err := ExtractNtfyTopicFromMQTT(config.MQTT.Topic, topic)
			if err != nil {
				logger.Error("Failed to extract Ntfy topic from MQTT topic", "error", err, "subscription", config.MQTT.Topic, "received", topic)
				return
			}

			// Build the dynamic Ntfy URL
			ntfyURL, err = BuildNtfyURL(config.Ntfy.URL, ntfyTopic)
			if err != nil {
				logger.Error("Failed to build Ntfy URL", "error", err, "base_url", config.Ntfy.URL, "ntfy_topic", ntfyTopic)
				return
			}

			logger.Info("Using dynamic Ntfy topic", "mqtt_topic", topic, "ntfy_topic", ntfyTopic, "ntfy_url", ntfyURL)
		} else {
			// Use the configured Ntfy URL directly for non-wildcard subscriptions
			ntfyURL = config.Ntfy.URL
		}

		// Forward to Ntfy with retry logic using cleaned message and extracted priority
		if err := ForwardToNtfy(ntfyURL, cleanedMessage, config.Ntfy.AuthToken, messagePriority, ntfyConfig, logger); err != nil {
			logger.Error("Failed to forward message to Ntfy after retries", "error", err)
		} else {
			logger.Info("Message forwarded to Ntfy successfully", "priority", messagePriority)
		}
	})
	if err != nil {
		logger.Error("Failed to connect to MQTT", "error", err)
		os.Exit(1)
	}
	defer mqttHandler.Disconnect(1000)

	logger.Info("Connected to MQTT broker and subscribed to topic", "topic", config.MQTT.Topic)

	// Initialize heartbeat if configured
	var hb heartbeat.Heartbeat
	if config.Heartbeat.URL != "" || config.Heartbeat.Port > 0 {
		var err error
		hb, err = heartbeat.NewHeartbeat(&heartbeat.Config{
			HeartbeatInterval: config.GetHeartbeatInterval(),
			LivenessThreshold: config.GetHeartbeatLivenessThreshold(),
			HeartbeatURL:      config.Heartbeat.URL,
			Port:              config.Heartbeat.Port,
			OnError: func(err error) {
				logger.Error("Heartbeat error", "error", err)
			},
		})
		if err != nil {
			logger.Error("Failed to create heartbeat", "error", err)
			os.Exit(1)
		}

		hb.Start()
		if config.Heartbeat.Port > 0 {
			logger.Info("Heartbeat server started", "port", config.Heartbeat.Port)
		} else {
			logger.Info("Heartbeat client started", "url", config.Heartbeat.URL, "interval", config.Heartbeat.Interval)
		}

		hb.Alive(time.Now())
		// Start a ticker to send heartbeats periodically while connected
		ticker := time.NewTicker(config.GetHeartbeatInterval())
		go func() {
			for range ticker.C {
				if hb != nil {
					hb.Alive(time.Now())
				}
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Info("Received shutdown signal, disconnecting from MQTT")
	mqttHandler.Disconnect(1000)
	logger.Info("Shutdown complete")
}
