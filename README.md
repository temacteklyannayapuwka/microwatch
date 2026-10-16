# MicroWatch

A lightweight service health monitor written in Go.

MicroWatch is being built to check HTTP endpoints and help locate unavailable services.

## Development

Requires Go 1.22 or newer. Uses the Go standard library.

```sh
go run ./cmd/microwatch
go build -o bin/microwatch ./cmd/microwatch
```

## Roadmap

- [x] Initialize the Go project
- [ ] Load and validate service configuration
- [ ] Implement HTTP health checks
- [ ] Schedule concurrent checks and track service state
- [ ] Expose status and health endpoints
- [ ] Add CI and release v0.1.0

## License

MIT
