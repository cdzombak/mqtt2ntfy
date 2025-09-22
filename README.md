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

heartbeat:
  url: "https://uptimekuma.example.com:9001/api/push/1234abcd?status=up&msg=OK&ping="  # Optional: URL for sending heartbeats
  interval: "30s"  # Optional: Interval between heartbeats (default: 30s)
  liveness_threshold: "60s"  # Optional: Liveness threshold (default: 60s)
  port: 8888  # Optional: Port for health endpoint server (if set > 0, server will start)
```
**Note:** Heartbeat functionality is automatically enabled when either `url` or `port` is configured.

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

## Wildcard Topic Support

mqtt2ntfy supports MQTT one-level wildcard topics (ending with `/#`). When subscribing to a wildcard topic, the last part of the received MQTT topic will be used as the Ntfy topic name.

### How Wildcard Topics Work

- **Subscribe to**: `my/notifications/#`
- **Receive message on**: `my/notifications/alerts` → **Send to Ntfy topic**: `alerts`
- **Receive message on**: `my/notifications/warnings` → **Send to Ntfy topic**: `warnings`

### Configuration

For wildcard topics, set the **Ntfy URL to the base URL** (without the specific topic):

```yaml
mqtt:
  topic: "my/notifications/#"
ntfy:
  url: "https://ntfy.sh"  # Base URL only - topic will be appended dynamically
```

Or with command-line flags:
```bash
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "home/sensors/#" --ntfy-url "https://ntfy.sh"
```

### Wildcard Examples

**Home automation sensors:**
```bash
# Subscribe to: home/sensors/#
# Messages on home/sensors/temperature → https://ntfy.sh/temperature
# Messages on home/sensors/humidity → https://ntfy.sh/humidity
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "home/sensors/#" --ntfy-url "https://ntfy.sh"
```

**Server monitoring:**
```bash
# Subscribe to: monitoring/#
# Messages on monitoring/cpu → https://alerts.example.com/cpu
# Messages on monitoring/disk → https://alerts.example.com/disk
mqtt2ntfy --mqtt-broker monitor.local --mqtt-topic "monitoring/#" --ntfy-url "https://alerts.example.com"
```

**Root wildcard:**
```bash
# Subscribe to: # (matches any single-level topic)
# Messages on alerts → https://ntfy.sh/alerts
# Messages on warnings → https://ntfy.sh/warnings
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "#" --ntfy-url "https://ntfy.sh"
```

### Limitations

- Only **one-level wildcards** (`/#`) are supported
- Multi-level wildcards (`/+` or nested levels) are not supported
- Received topics with multiple levels beyond the wildcard pattern will be rejected

## Message Priority Prefixes

mqtt2ntfy supports setting message priority using prefixes at the beginning of MQTT messages. If a message begins with a priority prefix followed by a pipe (`|`), the priority will be extracted and applied to the Ntfy notification, and the prefix will be stripped from the message content.

### Supported Priority Prefixes

- **`1|`** through **`5|`** - Set specific Ntfy priority levels (1=min/quiet, 5=max/urgent)

### How Priority Prefixes Work

The priority prefix is removed from the message before sending to Ntfy:

- **Message**: `2|Low priority info` → **Ntfy receives**: `Low priority info` with priority `2`
- **Message**: `5|Critical database error` → **Ntfy receives**: `Critical database error` with priority `5`
- **Message**: `Regular message` → **Ntfy receives**: `Regular message` with configured default priority

### Priority Examples

**Urgent priority alert:**
```bash
# Send message "5|Critical system failure" to MQTT
# Results in Ntfy notification with priority 5 (urgent) and message "Critical system failure"
mosquitto_pub -h localhost -t "alerts" -m "5|Critical system failure"
```

**Quiet priority notification:**
```bash
# Send message "1|Daily backup completed" to MQTT
# Results in Ntfy notification with priority 1 (quiet) and message "Daily backup completed"
mosquitto_pub -h localhost -t "alerts" -m "1|Daily backup completed"
```

**Combined with wildcard topics:**
```bash
# Subscribe to: monitoring/#
# Message "4|High CPU usage" on monitoring/servers → https://ntfy.sh/servers with priority 4
# Message "1|Cleanup complete" on monitoring/maintenance → https://ntfy.sh/maintenance with priority 1
mqtt2ntfy --mqtt-broker localhost --mqtt-topic "monitoring/#" --ntfy-url "https://ntfy.sh"
```

## Installation

### Debian via apt repository

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

### Homebrew

```shell
brew install cdzombak/oss/mqtt2ntfy
```

### Manual from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/mqtt2ntfy/releases). Debian packages for each release are available as well.

## Author

Chris Dzombak

- [dzombak.com](https://www.dzombak.com)
- [GitHub @cdzombak](https://github.com/cdzombak)

## License

GNU GPL v3; see [LICENSE](LICENSE) in this repo.
