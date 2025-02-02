package common

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
	"time"
)

var FRAMES_KEEPALIVE = []byte{
	0x12, 0x00, 0x00, 0x03, 0xe8, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x00,
}

type ConnectionDetails struct {
	// The TCP host to connect to
	Host string
	// The TCP port to connect to
	Port string
	// The client ID to use for the connection
	ClientId int
	// The connection ID to use for the connection
	ConnectionId string
}

// TCPStream connects to the liveview server using a TCP connection.
// Returns an error if the connection fails or if the stream ends unexpectedly.
// TODO: Support audio I/O
// TODO: Support command I/O (e.g. PTZ commands)
//
// ctx: the context to use for the stream
//
// connInfo: the connection details to use to connect to the liveview server
//
// writer: the pipe to write the stream data to
//
// Example: TCPStream(ctx, ConnectionDetails{Host: "example.com", Port: "443", ConnectionId: 1234, ClientId: 5678})
func TCPStream(ctx context.Context, connInfo ConnectionDetails, writer io.Writer) error {
	log.Printf("Connecting to %s:%s\n", connInfo.Host, connInfo.Port)

	client, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", connInfo.Host, connInfo.Port), &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         connInfo.Host,
		Certificates:       []tls.Certificate{},
	})
	if err != nil {
		return fmt.Errorf("unable to initialize stream: %w", err)
	} else {
		log.Println("Connected to", client.RemoteAddr())
	}
	defer client.Close()

	start := time.Now()
	frames := GetTCPAuthFrames(connInfo.ConnectionId, connInfo.ClientId)
	for _, frame := range frames {
		if _, err := client.Write(frame); err != nil {
			return fmt.Errorf("error sending connection header: %w", err)
		}
	}

	buf := make([]byte, 64)
	var streamErr error
stream:
	for {
		select {
		case <-ctx.Done():
			log.Println("Closing stream")
			break stream
		default:
			if err := client.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
				streamErr = fmt.Errorf("error setting read deadline: %w", err)
				break stream
			}

			n, err := client.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					streamErr = fmt.Errorf("connection closed gracefully by peer: %w", err)
				} else if errors.Is(err, syscall.ECONNRESET) {
					streamErr = fmt.Errorf("connection reset by peer: %w", err)
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					streamErr = fmt.Errorf("read timeout: %w", err)
				} else {
					streamErr = fmt.Errorf("error reading from server: %w", err)
				}
				break stream
			}

			if _, err := writer.Write(buf[:n]); err != nil {
				streamErr = fmt.Errorf("error writing to writer: %w", err)
				break stream
			}

			// Send a keep-alive ping to the server
			if time.Since(start) > time.Second {
				if err := sendPing(client); err != nil {
					streamErr = fmt.Errorf("error sending keep-alive: %w", err)
					break stream
				}

				// Reset the timer
				start = time.Now()
			}
		}
	}

	return streamErr
}

// sendPing sends a keep-alive ping to the server.
//
// client: the client connection to send the ping on
//
// Example: sendPing(client) = nil
func sendPing(client *tls.Conn) (err error) {
	if err := client.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return fmt.Errorf("error setting write deadline: %w", err)
	}

	if _, err := client.Write(FRAMES_KEEPALIVE); err != nil {
		return fmt.Errorf("error sending keep-alive: %w", err)
	}

	return nil
}
