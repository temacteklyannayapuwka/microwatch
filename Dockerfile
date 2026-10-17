FROM golang:1.26.9-alpine3.24 AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/microwatch ./cmd/microwatch

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/microwatch /microwatch
COPY configs/services.json /app/configs/services.json
WORKDIR /app
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/microwatch"]
CMD ["-interval", "10s", "-listen", "0.0.0.0:8080"]
