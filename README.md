# mqtt2ntfy

A simple program that subscribes to an MQTT topic and forwards messages to an [Ntfy](https://ntfy.sh) server.

## Usage

### Basic Usage
```bash
mqtt2ntfy --config config.yaml [--verbose]
```

## Configuration

mqtt2ntfy supports three configuration methods:

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

## Debian via apt repository

Set up my `oss` apt repository:

```shell
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/keys/dist-cdzombak-net.gpg -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 644 /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo mkdir -p /etc/apt/sources.list.d
sudo curl -fsSL https://dist.cdzombak.net/cdzombak-oss.sources -o /etc/apt/sources.list.d/cdzombak-oss.sources
sudo chmod 644 /etc/apt/sources.list.d/cdzombak-oss.sources
sudo apt update
```

Then install `mqtt2ntfy` via `apt-get`:

```shell
sudo apt-get install mqtt2ntfy
```

## Homebrew

```shell
brew install cdzombak/oss/mqtt2ntfy
```

## Manual from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/mqtt2ntfy/releases). Debian packages for each release are available as well.

## Author

Chris Dzombak

- [dzombak.com](https://www.dzombak.com)
- [GitHub @cdzombak](https://github.com/cdzombak)

## License

GNU GPL v3; see [LICENSE](LICENSE) in this repo.
