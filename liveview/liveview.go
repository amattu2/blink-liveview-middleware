package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var baseAPIURL = fmt.Sprintf("https://rest-%s.immedia-semi.com", config.Region)

func startLiveview() (map[string]interface{}, error) {
	headers := map[string]string{
		"app-build":         "ANDROID_28799573",
		"user-agent":        "37.0ANDROID_28799573",
		"locale":            "en_US",
		"x-blink-time-zone": "America/New_York",
		"token-auth":        config.Token,
		"content-type":      "application/json; charset=UTF-8",
	}

	jsonData := map[string]string{
		"intent": "liveview",
	}

	jsonValue, _ := json.Marshal(jsonData)
	url := fmt.Sprintf("%s/api/v2/accounts/%d/networks/%d/owls/%d/liveview", baseAPIURL, config.AccountID, config.NetworkID, config.CameraID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Initiated live view command: %s\n", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return result, nil
}

func pollCommand(networkId int, commandId int, interval int) {
	headers := map[string]string{
		"app-build":         "ANDROID_28799573",
		"user-agent":        "37.0ANDROID_28799573",
		"locale":            "en_US",
		"x-blink-time-zone": "America/New_York",
		"token-auth":        config.Token,
		"content-type":      "application/json; charset=UTF-8",
	}

	url := fmt.Sprintf("%s/api/v2/accounts/%d/networks/%d/commands/%d", baseAPIURL, config.AccountID, networkId, commandId)

	for {
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			return
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		client := &http.Client{Timeout: time.Second * 10}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func tcpStream(host string, port string, clientId int, connId string) {
	clientIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(clientIDBytes, uint32(clientId))

	connIDBytes := []byte(connId)

	connHeader := append([]byte{
		0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}, clientIDBytes...)

	connHeader = append(connHeader, []byte{
		0x01, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x10,
	}...)

	connHeader = append(connHeader, connIDBytes...)
	connHeader = append(connHeader, []byte{
		0x00, 0x00, 0x00, 0x01, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}...)

	keepAlive := []byte{
		0x12, 0x00, 0x00, 0x03, 0xe8, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x00,
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	ssock := tls.Client(conn, config)
	err = ssock.Handshake()
	if err != nil {
		fmt.Println("TLS handshake failed:", err)
		return
	}
	defer ssock.Close()

	start := time.Now()
	_, err = ssock.Write(connHeader)
	if err != nil {
		fmt.Println("Error sending connection header:", err)
		return
	}

	ffplayCmd := exec.Command("ffplay", "-f", "mpegts", "-err_detect", "ignore_err", "-")
	ffplayIn, err := ffplayCmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating ffplay stdin pipe:", err)
		return
	}

	err = ffplayCmd.Start()
	if err != nil {
		fmt.Println("Error starting ffplay:", err)
		return
	}
	defer ffplayCmd.Process.Kill()

	buf := make([]byte, 64)
	for {
		n, err := ssock.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from socket:", err)
			return
		}

		_, err = ffplayIn.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing to ffplay stdin:", err)
			return
		}

		if time.Since(start) > time.Second {
			_, err = ssock.Write(keepAlive)
			if err != nil {
				fmt.Println("Error sending keep-alive:", err)
				return
			}
			start = time.Now()
		}
	}
}

func livestream() {
	liveview, err := startLiveview()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	commandId := int(liveview["command_id"].(float64))
	pollingInterval := int(liveview["polling_interval"].(float64))

	fmt.Printf("Command ID: %d\n", commandId)
	fmt.Printf("Polling Interval: %d\n", pollingInterval)

	go pollCommand(config.NetworkID, commandId, pollingInterval)

	server := liveview["server"].(string)
	host := strings.Split(strings.Split(server, "/")[2], ":")[0]
	connID := strings.Split(strings.Split(server, "/")[len(strings.Split(server, "/"))-1], "_")[0]
	clientID, err := strconv.Atoi(strings.Split(server, "?client_id=")[1])
	if err != nil {
		fmt.Println("Error converting client_id to int:", err)
		return
	}

	fmt.Println("Host:", host)
	fmt.Println("ConnID:", connID)
	fmt.Println("ClientID:", clientID)

	tcpStream(host, "443", clientID, connID)
}

func main() {
	livestream()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
}
