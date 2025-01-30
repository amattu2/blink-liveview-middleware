package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"
)

func tcpStream(connInfo ConnectionDetails) {
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

	start := time.Now()
	// TODO: I think the header is invalid and forces the connection to close
	_, err = client.Write(GetTCPConnectionHeader(connInfo.ConnectionId, connInfo.ClientId))
	if err != nil {
		fmt.Println("Error sending connection header:", err)
		return
	} else {
		fmt.Println("Connection header sent")
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

	buf := make([]byte, 64)
	for {
		fmt.Println("Reading from socket...")

		err = client.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline:", err)
			break
		}

		n, err := client.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Socket closed by remote host")
				break
			}
			fmt.Println("Error reading from socket:", err)
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

	fmt.Println("Done...")
}

func Livestream(region string, token string, deviceType string, accountId int, networkId int, cameraId int) {
	// Tell Blink we want to start a liveview session
	liveViewPath, err := GetLiveviewPath(deviceType)
	if err != nil {
		fmt.Println(err)
		return
	}

	baseUrl := GetApiUrl(region)
	liveview, err := BeginLiveview(fmt.Sprintf(liveViewPath, baseUrl, accountId, networkId, cameraId), token)
	if err != nil {
		fmt.Println(err)
		return
	} else if liveview == nil || liveview.CommandId == 0 {
		fmt.Println("Error sending liveview command", liveview)
		return
	}

	// Poll the liveview command to keep the connection alive
	pollCtx, cancelCtx := context.WithCancel(context.Background())
	go PollCommand(pollCtx, fmt.Sprintf("%s/network/%d/command/%d", baseUrl, networkId, liveview.CommandId), token, liveview.PollingInterval)
	defer cancelCtx()

	// Stop the liveview session
	defer StopLiveview(fmt.Sprintf("%s/network/%d/command/%d/done", baseUrl, networkId, liveview.CommandId), token)

	// Connect to the liveview server
	connectionDetails, err := ParseConnectionString(liveview.Server)
	if err != nil {
		fmt.Println(err)
		return
	}

	tcpStream(*connectionDetails)
}
