package common

import (
	"context"
	"fmt"
)

// func tcpStream(connInfo ConnectionDetails) {
// 	clientIDBytes := make([]byte, 4)
// 	binary.BigEndian.PutUint32(clientIDBytes, uint32(connInfo.ClientId))

// 	connIDBytes := []byte(connInfo.ConnectionId)

// 	connHeader := append([]byte{
// 		0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	}, clientIDBytes...)

// 	connHeader = append(connHeader, []byte{
// 		0x01, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x10,
// 	}...)

// 	connHeader = append(connHeader, connIDBytes...)
// 	connHeader = append(connHeader, []byte{
// 		0x00, 0x00, 0x00, 0x01, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	}...)

// 	keepAlive := []byte{
// 		0x12, 0x00, 0x00, 0x03, 0xe8, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
// 		0x00,
// 	}

// 	addr := fmt.Sprintf("%s:%s", connInfo.Host, connInfo.Port)
// 	conn, err := net.Dial("tcp", addr)
// 	if err != nil {
// 		fmt.Println("Error connecting:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	config := &tls.Config{
// 		InsecureSkipVerify: true,
// 		ServerName:         connInfo.Host,
// 	}

// 	ssock := tls.Client(conn, config)
// 	err = ssock.Handshake()
// 	if err != nil {
// 		fmt.Println("TLS handshake failed:", err)
// 		return
// 	}
// 	defer ssock.Close()

// 	start := time.Now()
// 	_, err = ssock.Write(connHeader)
// 	if err != nil {
// 		fmt.Println("Error sending connection header:", err)
// 		return
// 	}

// 	ffplayCmd := exec.Command("ffplay", "-f", "mpegts", "-err_detect", "ignore_err", "-")
// 	ffplayIn, err := ffplayCmd.StdinPipe()
// 	if err != nil {
// 		fmt.Println("Error creating ffplay stdin pipe:", err)
// 		return
// 	}

// 	err = ffplayCmd.Start()
// 	if err != nil {
// 		fmt.Println("Error starting ffplay:", err)
// 		return
// 	}
// 	defer ffplayCmd.Process.Kill()

// 	buf := make([]byte, 64)
// 	for {
// 		n, err := ssock.Read(buf)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			fmt.Println("Error reading from socket:", err)
// 			return
// 		}

// 		_, err = ffplayIn.Write(buf[:n])
// 		if err != nil {
// 			fmt.Println("Error writing to ffplay stdin:", err)
// 			return
// 		}

// 		if time.Since(start) > time.Second {
// 			_, err = ssock.Write(keepAlive)
// 			if err != nil {
// 				fmt.Println("Error sending keep-alive:", err)
// 				return
// 			}
// 			start = time.Now()
// 		}
// 	}
// }

func Livestream(region string, token string, deviceType string, accountId int, networkId int, cameraId int) {
	// Tell Blink we want to start a liveview session
	liveViewPath, err := GetLiveviewPath(deviceType)
	if err != nil {
		fmt.Println(err)
		return
	}

	baseUrl := GetApiUrl(region)
	liveViewUrl := fmt.Sprintf(liveViewPath, baseUrl, accountId, networkId, cameraId)
	liveview, err := BeginLiveview(liveViewUrl, token)
	if err != nil {
		fmt.Println(err)
		return
	}
	if liveview == nil || liveview.CommandId == 0 {
		fmt.Println("Error starting liveview: invalid response", liveview)
		return
	}

	// Poll the liveview command to keep the connection alive
	pollCtx, cancelCtx := context.WithCancel(context.Background())
	pollCommandUrl := fmt.Sprintf("%s/network/%d/command/%d", baseUrl, networkId, liveview.CommandId)
	go PollCommand(pollCtx, pollCommandUrl, token, liveview.PollingInterval)
	defer cancelCtx()

	// Stop the liveview session
	defer StopLiveview(fmt.Sprintf("%s/network/%d/command/%d/done", baseUrl, networkId, liveview.CommandId), token)

	// Connect to the liveview server
	connection, err := ParseConnectionString(liveview.Server)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connecting to", connection.Host, connection.Port)
	// tcpStream(*connection)
	defer fmt.Println("Disconnected")
}
