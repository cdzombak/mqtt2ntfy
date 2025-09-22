# mqtt2ntfy

A simple program that subscribes to an MQTT topic and forwards messages to an [Ntfy](https://ntfy.sh) server.

## Usage

### Basic Usage
```bash
mqtt2ntfy --config config.yaml [--verbose]
```

### Using Command-Line Flags
```bash
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "sensors/temp" --ntfy-url "https://ntfy.sh/my-topic" [--verbose]
```

### Using Environment Variables (for containers)
```bash
export MQTT_USERNAME="user"
export MQTT_PASSWORD="pass"
export NTFY_AUTH_TOKEN="token"
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "sensors/temp" --ntfy-url "https://ntfy.sh/my-topic"
```

## Configuration

mqtt2ntfy supports three configuration methods with the following precedence:

**1. Command-line flags** (highest priority)
**2. Environment variables**
**3. Configuration file** (lowest priority)

### Configuration File

Copy `config.example.yaml` to `config.yaml` and update the values:

```yaml
mqtt:
  broker: "localhost"  # Protocol and port optional (defaults: tcp://, 1883)
  topic: "home/sensors/temperature"
  username: "your-mqtt-username"  # Optional
  password: "your-mqtt-password"  # Optional

ntfy:
  url: "https://ntfy.sh/your-topic-name"
  auth_token: "your-ntfy-auth-token"  # Optional
  priority: "3"  # Optional (1-5)
```

### Command-Line Flags

```bash
  --config string          Path to YAML configuration file (optional if all required flags provided)
  --verbose               Enable verbose logging
  --mqtt-broker string    MQTT broker URL (e.g., localhost, tcp://localhost:1883)
  --mqtt-topic string     MQTT topic to subscribe to
  --mqtt-username string  MQTT username for authentication
  --mqtt-password string  MQTT password for authentication
  --ntfy-url string       Ntfy server URL (e.g., https://ntfy.sh/your-topic)
  --ntfy-token string     Ntfy authentication token
  --ntfy-priority string  Ntfy message priority (1-5)
```

### Environment Variables

For sensitive values (recommended for containers):

```bash
MQTT_USERNAME          # MQTT username
MQTT_PASSWORD          # MQTT password
NTFY_AUTH_TOKEN        # Ntfy authentication token
```

### Configuration Examples

**Minimal with flags only:**
```bash
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "home/temp" --ntfy-url "https://ntfy.sh/alerts"
```

**With authentication:**
```bash
export MQTT_USERNAME="sensor_user"
export MQTT_PASSWORD="secret123"
export NTFY_AUTH_TOKEN="tk_abcd1234"
mqtt2ntfy --mqtt-broker mqtt.example.com --mqtt-topic "sensors/#" --ntfy-url "https://ntfy.sh/my-sensors"
```

**Mixed configuration (config file + overrides):**
```bash
# Uses config.yaml for defaults, overrides broker and adds auth via env vars
export NTFY_AUTH_TOKEN="tk_prod_token"
mqtt2ntfy --config config.yaml --mqtt-broker "ssl://prod-mqtt.example.com:8883"
```

## Installation

TK

## Author

Chris Dzombak

- [dzombak.com](https://www.dzombak.com)
- [GitHub @cdzombak](https://github.com/cdzombak)

## License

GNU GPL v3; see [LICENSE](LICENSE) in this repo.
