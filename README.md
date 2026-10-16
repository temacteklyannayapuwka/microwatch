# MicroWatch

A lightweight service health monitor written in Go.

MicroWatch is being built to check HTTP endpoints and help locate unavailable services.

## Development

Requires Go 1.22 or newer. Uses the Go standard library.

```sh
go run ./cmd/microwatch
go build -o bin/microwatch ./cmd/microwatch
```

## License

MIT
