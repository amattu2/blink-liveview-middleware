# Introduction

This project offers access into the liveview functionality of the Blink Smart
Security cameras. It has three entrypoints:

- A WebSocket service that can act as middleware between a web application and
the Blink Smart Security Camera
- A command to login to a Blink account and list the cameras available for liveview
- A command to watch the liveview stream from the command line (ffmpeg)

[![Go Report Card](https://goreportcard.com/badge/github.com/amattu2/blink-liveview-middleware)](https://goreportcard.com/report/github.com/amattu2/blink-liveview-middleware)
[![Test](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/test.yml/badge.svg)](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/test.yml)
[![CodeQL](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/codeql.yml/badge.svg)](https://github.com/amattu2/blink-liveview-middleware/actions/workflows/codeql.yml)

# Usage

See the following sections for more information on how to use each entry point.
These sections provide instructions on usage without compiling the code yourself.
If you would like to compile the code yourself, see the
Building From Source section below.

```bash
go run main.go <command> [flags]
```

> [!WARNING]
> Prior to running any of the commands, ensure that you have ffmpeg installed
> on your system, and configured in your PATH (if applicable).

## Account Command

The account command is a simple interface for interacting with the Blink Liveview
APIs without retrieving technical connection details on your own. Provide the email
address and password for the Blink account you wish to use, and the command will
perform the necessary steps to obtain the API token and account information.

Once retrieved, it will prompt you to select a camera to watch the liveview stream
from, and then open the liveview stream in a new window using ffplay.

Additionally, the command will output the API Token, Account ID, and Region, which
can be used to shortcut the process in the future.

```bash
go run main.go account \
    [--email=<email>] \
    [--token=<api token> --account-id=<account id> --region=<region>]
```

An explanation of the command line flags is provided below:

Option 1: Email & Password

- `-e`, `--email`: The email address of the Blink account to use

> [!NOTE]
> The password is not provided as a command line flag for security reasons.
> You will be prompted to enter the password after running the command.

Option 2: API Token, Account ID, & Region

- `-t`, `--token`: The API token for the current session. This is returned via
the Blink login flow
- `-a`, `--account-id`: The account ID of the Blink account
- `-r`, `--region`: The region of the Blink account (e.g. `u014`, `u011`, etc.)

## Liveview Command

The liveview command is a direct way to watch the liveview stream from a Blink
Smart Security Camera. It can be used in place of the above "account command"
if you already have all of the necessary information to connect to the Blink API.

This is made available primarily used for testing, but can be used as
a standalone tool if desired.

Upon running the command, you should see a new ffplay window open with the
liveview stream. The stream will be gracefully closed by terminating the CLI process.

```bash
go run main.go liveview \
  --region=<region> \
  --token=<api token> \
  --device-type=<device type> \
  --account-id=<account id> \
  --network-id=<network id> \
  --camera-id=<camera id>
```

An explanation of the command line flags is provided below:

- `-r`, `--region`: The region of the Blink account (e.g. `u014`, `u011`, etc.).
This is returned via the Blink login flow
- `-t`, `--token`: The API token for the current session. This is also returned via
the Blink login flow
- `-d`, `--device-type`: The type of (camera) device to connect to (e.g. `owl`, `doorbell`).
- `-a`, `--account-id`: The account ID of the Blink account
- `-n`, `--network-id`: The ID of the network that the camera is on
- `-c`, `--camera-id`: The ID of the camera to watch

## WebSocket Middleware

This section is broken down into two parts: the server and the client. The server
is a WebSocket service that acts as middleware between a web application and the
Blink Smart Security Camera. The client section provides an example of how to
interface with the WebSocket server using JavaScript.

Out of the box, the server provides a demo UI that can be used to test the liveview
stream via a web browser.

### Server Usage

The server is a basic Go HTTP server that utilizes the Gorilla WebSocket library.
It has no built-in authentication or knowledge of the Blink API (beyond liveview),
so it is entirely up to your implementing application to provide the necessary
information to the server.

Each client that connects to the WebSocket is independent
of the others, so you can have multiple streams running at the same time
without overlapping.

Start the server with the following command:

```bash
go run main.go server [--address=<addr>] [--env=<env>] [--origins=<origins>]
```

An explanation of the command line flags is provided below:

- `-a`, `--address`: The address to bind the server to (e.g. `:8080`)
- `-e`, `--env`: The environment to run the server in (`development`, `production`).
If `production` is specified, the demo UI will be disabled.
- `-o`, `--origins`: A comma-separated list of allowed WebSocket client origins.
By default, the current origin is allowed. Use `*` to allow all origins.

Then open the sample web application in your browser. Provide the necessary
authentication information on the demo UI and click the "Start Liveview" button:

<http://localhost:8080/index.html>

> [!NOTE]
> The server does not currently limit the maximum number of clients that can
> connect OR liveview at the same time. This may cause performance issues.

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

ws.onopen = () => {
    // Send the authentication information to the server
    // NOTE: This should be done within 8 seconds of connecting,
    // or the server will close the connection
    const data = JSON.stringify({
        command: "liveview:start",
        data: {
            // Refer to the liveview CLI arguments for details on these fields
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
        // Handle incoming video packets
        return;
    }

    const data = JSON.parse(evt.data);
    if (data?.command === "liveview:stop") {
        // The server stopped the liveview
        // Handle receipt of the stop command (e.g. stop the video player)
    } else if (data?.command === "liveview:start") {
        // The server opened the liveview
        // binary data will begin shortly (delay of about 5 seconds)
    }
};
```

Refer to the demo UI [source code](static/index.html) for a more detailed example
of how to connect and integrate the liveview stream into your web application.

## Building From Source

Using make, you can build the project for your platform. The Makefile
provides a simple interface for building the project, but you can also
build it manually using the `go build` command.

To build the project using make for all platforms, run the following command:

```bash
make build
```

To build a specific platform, use the `build_[platform]` target.

```bash
make build_[platform]
```

Platforms: `linux`, `mac`, `freebsd`, `windows`

To build the project manually, you can use the `go build` command. The

```bash
go build -a -o bin/file_name_here main.go
```

To clean the workspace and remove any generated files, run:

```bash
make clean
```

# Blink Liveview Process

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
- Gorilla WebSocket
- ffmpeg / ffplay
