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

The command validates the configuration and checks each service once, sequentially.
HTTP 2xx responses are reported as UP; other status codes and request errors are
reported as DOWN. Redirects are not followed. Output includes the HTTP status and
response time when a response is received.

The process exits with code 0 if all services are UP, or code 1 if configuration
loading fails or any service is DOWN.

## Tests

```sh
go test ./...
```

## License

MIT
