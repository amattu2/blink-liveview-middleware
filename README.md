# Introduction

This project offers access into the liveview functionality of the Blink Smart
Security cameras. It has two entrypoints:

- A WebSocket service that can act as middleware between a web application and
the Blink Smart Security Camera
- A CLI command to watch the liveview stream from the command line (ffmpeg)

[![Go Report Card](https://goreportcard.com/badge/github.com/amattu2/blink-liveview-middleware)](https://goreportcard.com/report/github.com/amattu2/blink-liveview-middleware)
[![Test](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/test.yml/badge.svg)](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/test.yml)
[![CodeQL](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/codeql.yml/badge.svg)](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/codeql.yml)

# Usage

See the following sections for more information on how to use each entry point.
These sections provide instructions on usage without compiling the code yourself.
If you would like to compile the code yourself, see the
Building From Source section below.

## Liveview CLI Usage

The liveview CLI tool is a simple command line tool that uses ffmpeg to watch the
liveview stream from a Blink Smart Security Camera. It's primarily used for testing,
but can be used as a standalone tool if desired.

```bash
go run cmd/cli/main.go \
  --region=<region> \
  --token=<api token> \
  --device-type=<lotus|owl|doorbell|etc> \
  --account-id=<account id> \
  --network-id=<network id> \
  --camera-id=<camera id>
```

## WebSocket Server Usage

The WebSocket server is a simple middleware service that can be used to proxy
the liveview stream from a Blink Smart Security Camera to a web application. It
is designed to be used in conjunction with a web application that can send
commands to the server to start and stop the liveview stream.

It has no built-in authentication or knowledge of the Blink API, so it is up to
your implementing application to provide the necessary information to the server.

Start the server with the following command:

```bash
TODO
```

Then open the example web application in your browser. Provide the necessary
authentication information on the demo UI and click the "Start Liveview" button:

<http://localhost:8080/index.html>

When deploying the service to production,
you should ensure that the demo webpage is disabled.

## Building From Source

```bash
TODO
```

# Requirements & Dependencies

- Go 1.23+
- ffmpeg (for the CLI tool)
