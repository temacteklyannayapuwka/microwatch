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

The command validates the configuration and prints the configured services. It does not send health-check requests yet.

## License

MIT
