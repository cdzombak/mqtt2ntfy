package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var configPath string
	var verbose bool

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

		// Forward to Ntfy with retry logic
		if err := ForwardToNtfy(config.Ntfy.URL, string(payload), config.Ntfy.AuthToken, config.Ntfy.Priority, ntfyConfig); err != nil {
			logger.Error("Failed to forward message to Ntfy after retries", "error", err)
		} else {
			logger.Info("Message forwarded to Ntfy successfully")
		}
	})
	if err != nil {
		logger.Error("Failed to connect to MQTT", "error", err)
		os.Exit(1)
	}
	defer mqttHandler.Disconnect(1000)

	logger.Info("Connected to MQTT broker and subscribed to topic", "topic", config.MQTT.Topic)

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Info("Received shutdown signal, disconnecting from MQTT")
	mqttHandler.Disconnect(1000)
	logger.Info("Shutdown complete")
}
