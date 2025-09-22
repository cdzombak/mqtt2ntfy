package main

import (
	"context"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient interface for dependency injection and testing
type MQTTClient interface {
	Connect() mqtt.Token
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token
	Disconnect(quiesce uint)
}

// MQTTHandler wraps the paho MQTT client
type MQTTHandler struct {
	client mqtt.Client
}

// NewMQTTHandler creates a new MQTT handler with configurable timeouts
func NewMQTTHandler(broker, username, password string, connectTimeout, pingTimeout time.Duration) (*MQTTHandler, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("mqtt2ntfy")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(pingTimeout)
	opts.SetConnectTimeout(connectTimeout)

	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	return &MQTTHandler{client: client}, nil
}

// Connect implements MQTTClient interface
func (m *MQTTHandler) Connect() mqtt.Token {
	return m.client.Connect()
}

// Subscribe implements MQTTClient interface
func (m *MQTTHandler) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return m.client.Subscribe(topic, qos, callback)
}

// Disconnect implements MQTTClient interface
func (m *MQTTHandler) Disconnect(quiesce uint) {
	m.client.Disconnect(quiesce)
}

// ConnectAndSubscribe connects to MQTT broker and subscribes to topic with retry logic
func ConnectAndSubscribe(ctx context.Context, broker, topic, username, password string, connectTimeout, pingTimeout time.Duration, messageHandler func(string, []byte)) (*MQTTHandler, error) {
	handler, err := NewMQTTHandler(broker, username, password, connectTimeout, pingTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create MQTT handler: %w", err)
	}

	// Retry connection up to 3 times
	for i := range 3 {
		token := handler.Connect()
		if token.Wait() && token.Error() == nil {
			break
		}
		if i == 2 {
			return nil, fmt.Errorf("failed to connect to MQTT broker after 3 attempts: %w", token.Error())
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	token := handler.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		messageHandler(msg.Topic(), msg.Payload())
	})
	if token.Wait() && token.Error() != nil {
		handler.Disconnect(0)
		return nil, fmt.Errorf("failed to subscribe to topic: %w", token.Error())
	}

	return handler, nil
}
