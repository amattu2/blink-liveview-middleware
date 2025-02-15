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

## Liveview Command

The liveview script is a simple tool that uses ffmpeg to watch the
liveview stream from a Blink Smart Security Camera. It's primarily used for testing,
but can be used as a standalone tool if desired.

```bash
go run cmd/liveview/main.go \
  --region=<region> \
  --token=<api token> \
  --device-type=<lotus|owl|doorbell|etc> \
  --account-id=<account id> \
  --network-id=<network id> \
  --camera-id=<camera id>
```

## WebSocket Middleware

### Server Usage

The WebSocket server is a middleware service that can be used to proxy
liveview streams from a Blink Smart Security Camera to a web application. It
is designed to be used in conjunction with a web application that can send
commands to the server to start and stop the liveview stream.

It has no built-in authentication or knowledge of the Blink API, so it is up to
your implementing application to provide the necessary information to the server.
Each client that connects to the WebSocket is independent of the others, so
you can have multiple streams running at the same time without overlapping.

Start the server with the following command:

```bash
go run cmd/server/main.go --address=:8080 --env=<development|production>
```

Then open the sample web application in your browser. Provide the necessary
authentication information on the demo UI and click the "Start Liveview" button:

<http://localhost:8080/index.html>

When deploying the service to production, this page is disabled by default.

### Client Usage

Each client that connects to the WebSocket server is independent of the others,
which means that each client must forward the Blink authentication information
to the server once connected.

By default, the server will close the connection if the client does not start
liveview or send some sort of command within `8 seconds` of connecting.

The following is an example of how to connect to the WebSocket server using
JavaScript:

```javascript
// Open a WebSocket connection to the server
const ws = new WebSocket('ws://localhost:8080/liveview');
ws.binaryType = "arraybuffer";

// Send the authentication information to the server
// NOTE: This does not have to be done immediately after opening the connection
ws.onopen = () => {
    const data = JSON.stringify({
        command: "liveview:start",
        data: {
          account_region: "",
          api_token: "",
          account_id: "",
          network_id: "",
          camera_id: "",
          camera_type: "",
        },
    });

    ws.send(data);
};

// Handle incoming messages from the server
ws.onmessage = (evt) => {
    if (evt.data instanceof ArrayBuffer) {
        // Handle incoming livestream packets
        return;
    }

    const data = JSON.parse(evt.data);
    if (data?.command === "liveview:stop") {
        // Handle receipt of the stop command
        // The server stopped the livestream
    } else if (data?.command === "liveview:start") {
        // The server opened the livestream
        // binary data will begin shortly
    }
};
```

## Building From Source

```bash
# Server binary (Windows)
go build -a -o bin/server.exe cmd/server/main.go
# Liveview binary (Windows)
go build -a -o bin/liveview.exe cmd/liveview/main.go
```

# Liveview Process

The general process behind obtaining a liveview stream from a Blink camera is
outlined below, ignoring the specifics of the Blink API and any potential error states.

```mermaid
---
title: Blink Smart Security Liveview Process
---
sequenceDiagram
    participant C as Client (You)
    participant B as Blink HTTP API
    participant T as Blink TCP Server

    C->>B: POST /liveview
    B->>C: Liveview response 
    Note over B,C: Returns TCP server and credentials
    par TCP Connection
        C->>T: Open TCP connection
        C->>T: TCP Auth Frame (1/5)
        C->>T: TCP Auth Frame (2/5)
        C->>T: TCP Auth Frame (3/5)
        C->>T: TCP Auth Frame (4/5)
        C->>T: TCP Auth Frame (5/5)
        loop
            T->>C: Binary stream data
        end
    and
        loop
            C->>B: POST /command status
            B->>C: Command Response
        end
    end
    C->>B: POST /command/done
    Note over B,C: Sent once the TCP connection is closed
    B->>C: Command Response
```

# Dependencies

- Go 1.23+
- ffmpeg / ffplay
