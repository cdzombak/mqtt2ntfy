package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var configPath string
	var verbose bool

	flag.StringVar(&configPath, "config", "", "Path to YAML configuration file (required)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	if configPath == "" {
		fmt.Println("Error: --config flag is required")
		flag.Usage()
		os.Exit(1)
	}

	logger := SetupLogger(verbose)
	logger.Info("Starting mqtt2ntfy", "config", configPath)

	config, err := LoadConfig(configPath)
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
