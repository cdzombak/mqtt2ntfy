# mqtt2ntfy

A simple program that subscribes to an MQTT topic and forwards messages to an [Ntfy](https://ntfy.sh) server.

## Usage

```bash
mqtt2ntfy --config config.yaml [--verbose]
```

## Configuration

Copy `config.example.yaml` to `config.yaml` and update the values:

```yaml
mqtt:
  broker: "tcp://localhost:1883"
  topic: "home/sensors/temperature"
  username: "your-mqtt-username"  # Optional
  password: "your-mqtt-password"  # Optional

ntfy:
  url: "https://ntfy.sh/your-topic-name"
  auth_token: "your-ntfy-auth-token"  # Optional
  priority: "3"  # Optional (1-5)
```

## Installation

TK

## Author

Chris Dzombak

- [dzombak.com](https://www.dzombak.com)
- [GitHub @cdzombak](https://github.com/cdzombak)

## License

GNU GPL v3; see [LICENSE](LICENSE) in this repo.
