package common

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"syscall"
	"time"
)

// TCPStream connects to the liveview server using a TCP connection
// TODO: Support multiple output methods (e.g. ffmpeg, ffplay, etc.)
// TODO: Support audio I/O
// TODO: Support command I/O (e.g. PTZ commands)
//
// connInfo: the connection details to use to connect to the liveview server
//
// Example: TCPStream(ConnectionDetails{Host: "example.com", Port: "443", ConnectionId: 1234, ClientId: 5678})
func TCPStream(connInfo ConnectionDetails) {
	fmt.Printf("Initializing stream to %s:%s\n", connInfo.Host, connInfo.Port)

	client, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", connInfo.Host, connInfo.Port), &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         connInfo.Host,
		Certificates:       []tls.Certificate{},
	})
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	} else {
		fmt.Println("Connected to", client.RemoteAddr())
	}
	defer client.Close()

	ffplayCmd := exec.Command("ffplay", "-f", "mpegts", "-err_detect", "ignore_err", "-")
	inputPipe, err := ffplayCmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating ffplay stdin pipe:", err)
		return
	}

	if err := ffplayCmd.Start(); err != nil {
		fmt.Println("Error starting ffplay:", err)
		return
	}
	defer ffplayCmd.Process.Kill()

	start := time.Now()
	frames := GetTCPAuthFrames(connInfo.ConnectionId, connInfo.ClientId)
	for _, frame := range frames {
		if _, err := client.Write(frame); err != nil {
			fmt.Println("Error sending connection header:", err)
			return
		}
	}

	buf := make([]byte, 64)
	for {
		if err := client.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			fmt.Println("Error setting read deadline:", err)
			break
		}

		n, err := client.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Connection closed gracefully by peer")
			} else if errors.Is(err, syscall.ECONNRESET) {
				fmt.Println("Connection reset by peer (ECONNRESET)")
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Read timeout, connection might be closed")
			} else {
				fmt.Println("Other read error:", err)
			}
			break
		}

		if _, err := inputPipe.Write(buf[:n]); err != nil {
			fmt.Println("Error writing to ffplay stdin:", err)
			break
		}

		// Send a keep-alive ping to the server
		if time.Since(start) > time.Second {
			if err := sendPing(client); err != nil {
				fmt.Println("Error sending keep-alive:", err)
				break
			}

			// Reset the timer
			start = time.Now()
		}
	}

	inputPipe.Close()

	if err := ffplayCmd.Wait(); err != nil {
		fmt.Println("Error waiting for ffplay:", err)
	}

	fmt.Println("Stream ended...")
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
