package common

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var FRAMES_KEEPALIVE = []byte{
	0x12, 0x00, 0x00, 0x03, 0xe8, 0x00, 0x00, 0x00,
	0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x00,
}

type TCPConnectionDetails struct {
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
// TODO: Support command I/O (e.g. PTZ commands)
//
// ctx: the context to use for the stream
//
// connInfo: the connection details to use to connect to the liveview server
//
// writer: the pipe to write the stream data to
//
// Example: TCPStream(ctx, "mock-server-path", writer) = nil
func TCPStream(ctx context.Context, server string, writer io.Writer) error {
	connInfo, err := parseConnectionString(server)
	if err != nil {
		return fmt.Errorf("error parsing connection string: %w", err)
	}

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
	defer log.Printf("Disconnected from %s:%s\n", connInfo.Host, connInfo.Port)

	start := time.Now()
	frames := connInfo.GenerateAuthFrames()
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

// parseConnectionString parses the connection string to extract the connection details
//
// url: the connection string to parse
//
// Example: parseConnectionString("TODO")
func parseConnectionString(server string) (*TCPConnectionDetails, error) {
	parsedUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	if parsedUrl.Hostname() == "" {
		return nil, fmt.Errorf("invalid host")
	}

	if parsedUrl.Port() != "443" {
		return nil, fmt.Errorf("unexpected port %s. Expecting 443", parsedUrl.Port())
	}

	pathSegments := strings.Split(parsedUrl.Path, "/")
	if len(pathSegments) == 0 {
		return nil, fmt.Errorf("invalid path")
	}

	connID := strings.Split(pathSegments[len(pathSegments)-1], "_")
	if len(connID) < 2 || connID[0] == "" {
		return nil, fmt.Errorf("invalid connection ID")
	}

	clientID, err := strconv.Atoi(parsedUrl.Query().Get("client_id"))
	if clientID == 0 || err != nil {
		return nil, fmt.Errorf("invalid client ID")
	}

	return &TCPConnectionDetails{
		Host:         parsedUrl.Hostname(),
		Port:         parsedUrl.Port(),
		ClientId:     clientID,
		ConnectionId: connID[0],
	}, nil
}

// GenerateAuthFrames returns the header payload for the TCP connection
//
// connectionId: the connection ID to use in the header
//
// clientId: the client ID to use in the header
//
// Example: GenerateAuthFrames("connection-id", 123)
func (details TCPConnectionDetails) GenerateAuthFrames() [][]byte {
	// Frame 1 (unknown)
	frame1 := []byte{
		0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// Frame 2 (Client ID)
	clientIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(clientIDBytes, uint32(details.ClientId))
	frame2 := []byte{
		clientIDBytes[0], clientIDBytes[1], clientIDBytes[2], clientIDBytes[3],
	}

	// Frame 3 (unknown)
	frame3 := []byte{
		0x01, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x10,
	}

	// Frame 4 (Connection ID)
	frame4 := []byte(details.ConnectionId)

	// Frame 5 (unknown)
	frame5 := []byte{
		0x00, 0x00, 0x00, 0x01, 0x0a, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	}

	return [][]byte{
		frame1,
		frame2,
		frame3,
		frame4,
		frame5,
	}
}
