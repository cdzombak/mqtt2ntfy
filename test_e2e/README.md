# End-to-End Testing with Docker Compose

This directory contains a complete Docker Compose stack for testing mqtt2ntfy end-to-end with zero configuration required.

## What's Included

- **Mosquitto MQTT Broker** - Eclipse Mosquitto with anonymous access enabled
- **Ntfy Server** - Self-hosted ntfy instance with web UI
- **mqtt2ntfy** - Built from current checkout, connects MQTT to Ntfy

## Quick Start

1. **Start the stack:**
   ```bash
   cd test_e2e
   docker-compose up
   ```

2. **Open Ntfy web UI:**
   - Visit: http://localhost:2080
   - Subscribe to topic: `mqtt2ntfy-test`

3. **Send a test message:**
   ```bash
   # In another terminal
   docker-compose exec mosquitto mosquitto_pub -h mosquitto -t test/messages -m "Hello from MQTT!"
   ```

4. **Verify:** The message should appear in the Ntfy web UI

## Alternative Testing Methods

### Using mosquitto_pub from host (if installed):
```bash
mosquitto_pub -h localhost -p 2083 -t test/messages -m "Test message from host"
```

### Using any MQTT client:
- **Broker:** localhost:2083
- **Topic:** test/messages
- **Message:** Any text content

## Services

- **MQTT Broker:** localhost:2083 (TCP), localhost:2081 (WebSocket)
- **Ntfy Web UI:** http://localhost:2080
- **Ntfy Topic:** mqtt2ntfy-test

## Configuration

All services are preconfigured and communicate via Docker networking:
- MQTT broker allows anonymous connections
- mqtt2ntfy subscribes to `test/messages` topic
- Messages are forwarded to `mqtt2ntfy-test` topic on Ntfy

## Cleanup

```bash
docker-compose down
docker-compose down -v  # Also remove volumes
```