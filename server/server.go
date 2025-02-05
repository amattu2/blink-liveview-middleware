package server

import (
	"blink-liveview-websocket/common"
	"context"
	"io"
	"log"
	"net/http"
	"os/exec"
	"slices"
	"strconv"

	"github.com/gorilla/websocket"
)

type MessageData struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{}

var VALID_COMMANDS = []string{
	"liveview:start",
	"liveview:stop",
}

// The buffer size before dispatching the data to the WebSocket connection in bytes
var BUFFER_SIZE = 12 * 1024

func liveviewHandler(ctx context.Context, c *websocket.Conn, data map[string]interface{}) {
	region := data["account_region"].(string)
	token := data["api_token"].(string)
	account_id, _ := strconv.Atoi(data["account_id"].(string))
	network_id, _ := strconv.Atoi(data["network_id"].(string))
	camera_id, _ := strconv.Atoi(data["camera_id"].(string))
	device_type := data["camera_type"].(string)

	// TODO: Strip metadata from the stream before piping it to the client
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-f", "mpegts",
		"-err_detect", "ignore_err",
		"pipe:1",
	)
	inputPipe, err := ffmpegCmd.StdinPipe()
	if err != nil {
		log.Println("error creating ffplay stdin pipe", err)
	}
	defer inputPipe.Close()

	outputPipe, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		log.Println("error creating ffplay stdout pipe", err)
	}
	defer outputPipe.Close()

	if err := ffmpegCmd.Start(); err != nil {
		log.Println("error starting ffplay", err)
	}
	defer ffmpegCmd.Process.Kill()

	// TODO: Handle Livestream errors and propagate them to the client
	go common.Livestream(ctx, common.AccountDetails{
		Region:     region,
		Token:      token,
		DeviceType: device_type,
		AccountId:  account_id,
		NetworkId:  network_id,
		CameraId:   camera_id,
	}, inputPipe)

	// Tell the client that the liveview has started
	c.WriteJSON(MessageData{
		Command: "liveview:start",
		Data: map[string]interface{}{
			"message": "Liveview started",
		},
	})

	// Forward messages from the pipe to the WebSocket connection
	go func() {
		buf := make([]byte, BUFFER_SIZE)
		tempBuf := make([]byte, 0, BUFFER_SIZE)

		// Read from the pipe and write to the WebSocket connection
		for {
			n, err := outputPipe.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from pipe: %v", err)
				}
				break
			}

			tempBuf = append(tempBuf, buf[:n]...)
			for len(tempBuf) >= BUFFER_SIZE {
				if err := c.WriteMessage(websocket.BinaryMessage, tempBuf[:BUFFER_SIZE]); err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					break
				}
				tempBuf = tempBuf[BUFFER_SIZE:]
			}
		}

		// Send any remaining data in tempBuf
		if len(tempBuf) > 0 {
			if err := c.WriteMessage(websocket.BinaryMessage, tempBuf); err != nil {
				log.Printf("Error writing remaining data to WebSocket: %v", err)
			}
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Tell the client that the liveview has stopped
	c.WriteJSON(MessageData{
		Command: "liveview:stop",
		Data: map[string]interface{}{
			"message": "Liveview stopped. Context cancelled",
		},
	})

	if err := ffmpegCmd.Wait(); err != nil {
		log.Println("error waiting for ffplay", err)
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	var ctx context.Context
	var cancelCtx context.CancelFunc
	var liveviewStarted bool
	for {
		var message MessageData
		if err := c.ReadJSON(&message); err != nil {
			log.Printf("read error: %v", err)
			break
		}

		if !slices.Contains(VALID_COMMANDS, message.Command) {
			log.Printf("invalid command: %v", message.Command)
			break
		} else if message.Command == "liveview:start" {
			log.Println("Client requested liveview:start")

			ctx, cancelCtx = context.WithCancel(context.Background())
			go liveviewHandler(ctx, c, message.Data)
			liveviewStarted = true
		} else if message.Command == "liveview:stop" && liveviewStarted {
			log.Println("Client requested liveview:stop")
			cancelCtx()
		}
	}

	defer cancelCtx()
}

func Run(address *string, env *string) {
	http.HandleFunc("/liveview", websocketHandler)

	if *env == "development" {
		log.Println("Enabled static file server")
		http.Handle("/", http.FileServer(http.Dir("./static")))
	}

	if err := http.ListenAndServe(*address, nil); err != nil {
		log.Fatal(err)
	}
}
