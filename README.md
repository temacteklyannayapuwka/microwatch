# MicroWatch

A lightweight service health monitor written in Go.

MicroWatch is being built to check HTTP endpoints and help locate unavailable services.

## Development

Requires Go 1.22 or newer. Uses the Go standard library.

```sh
go run ./cmd/microwatch
go build -o bin/microwatch ./cmd/microwatch
```

## Configuration

```sh
go run ./cmd/microwatch -config configs/services.json
```

The JSON file defines a positive `timeout` duration and a non-empty `services` list.
Each service needs a unique `name` and an absolute HTTP or HTTPS `url`.

The command validates the configuration and checks all services once, concurrently.
Results are printed in configuration order after all checks finish.
HTTP 2xx responses are reported as UP; other status codes and request errors are
reported as DOWN. Redirects are not followed. Output includes the HTTP status and
response time when a response is received.

In single-check mode, the process exits with code 0 if all services are UP, or
code 1 if configuration loading fails or any service is DOWN.

## Continuous monitoring

```sh
go run ./cmd/microwatch -interval 10s
```

The first round starts immediately. Each completed round is followed by the
specified pause, so rounds never overlap. DOWN results do not stop monitoring.
Ctrl+C or SIGTERM cancels active requests and stops the monitor. Without
`-interval`, the command keeps its single-check behavior. Negative intervals
are rejected.

Continuous monitoring also reports state changes with UTC timestamps. A transition
to DOWN is reported as `BECAME DOWN`; a transition to UP is reported as `RECOVERED`
with the observed downtime. Repeated checks in the same state do not produce
additional transition messages. The first check establishes each service's state.
Times reflect completed check rounds, so downtime is an estimate based on polling.
State is kept in memory and resets when the process restarts.

## Tests

```sh
go test ./...
```

## License

MIT
