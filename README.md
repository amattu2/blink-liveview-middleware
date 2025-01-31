# Introduction

A simple websocket middleware service to act as the intermediary between a web application
and the Blink Smart Security Camera livestream API.

[![Go Report Card](https://goreportcard.com/badge/github.com/amattu2/blink-liveview-middleware)](https://goreportcard.com/report/github.com/amattu2/blink-liveview-middleware)

# Usage

There are two entry points for this service, the CLI and the HTTP (websocket)
server. The CLI-based approach is designed to be used for testing and debugging purposes,
while the HTTP server is designed to be used in a production environment.

See the following sections for more information on how to use each entry point.

## CLI Usage

```bash
go run cmd/cli/main.go \
  --region=<region> \
  --token=<api token> \
  --device-type=<lotus|owl|doorbell|etc> \
  --account-id=<account id> \
  --network-id=<network id> \
  --camera-id=<camera id>
```

## HTTP Server Usage

```bash
TODO
```

# Requirements & Dependencies

TODO
