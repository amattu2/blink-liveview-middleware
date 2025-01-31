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
//
// connInfo: the connection details to use to connect to the liveview server
//
// Example: TCPStream(ConnectionDetails{Host: "example.com", Port: "443", ConnectionId: 1234, ClientId: 5678})
func TCPStream(connInfo ConnectionDetails) {
	fmt.Printf("Connecting to %s:%s\n", connInfo.Host, connInfo.Port)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", connInfo.Host, connInfo.Port))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	client := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         connInfo.Host,
	})

	err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		fmt.Println("Error setting read deadline:", err)
		return
	}
	err = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		fmt.Println("Error setting write deadline:", err)
		return
	}

	err = client.Handshake()
	if err != nil {
		fmt.Println("TLS handshake failed:", err)
		return
	} else {
		fmt.Println("TLS handshake successful")
	}
	defer client.Close()

	err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		fmt.Println("Error setting read deadline:", err)
		return
	}
	err = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		fmt.Println("Error setting write deadline:", err)
		return
	}

	ffplayCmd := exec.Command("ffplay", "-f", "mpegts", "-err_detect", "ignore_err", "-")
	ffplayIn, err := ffplayCmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating ffplay stdin pipe:", err)
		return
	} else {
		fmt.Println("Created ffplay stdin pipe")
	}

	err = ffplayCmd.Start()
	if err != nil {
		fmt.Println("Error starting ffplay:", err)
		return
	} else {
		fmt.Println("Started ffplay")
	}
	defer ffplayCmd.Process.Kill()

	start := time.Now()
	_, err = client.Write(GetTCPAuthFrame(connInfo.ConnectionId, connInfo.ClientId))
	if err != nil {
		fmt.Println("Error sending connection header:", err)
		return
	} else {
		fmt.Println("Connection header sent")
	}

	buf := make([]byte, 1024)
	for {
		fmt.Println("Reading from socket...")

		err = client.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
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

		if n == 0 {
			fmt.Println("No data read from socket")
			continue
		}

		fmt.Printf("Read %d bytes from socket\n", n)

		_, err = ffplayIn.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing to ffplay stdin:", err)
			break
		}

		if time.Since(start) > time.Second {
			_, err = client.Write(FRAMES_KEEPALIVE)
			if err != nil {
				fmt.Println("Error sending keep-alive:", err)
				break
			}
			start = time.Now()
		}
	}

	ffplayIn.Close()

	if err := ffplayCmd.Wait(); err != nil {
		fmt.Println("Error waiting for ffplay:", err)
	}

	fmt.Println("Stream ended...")
}
