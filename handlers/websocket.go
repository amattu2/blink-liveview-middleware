package handlers

import (
	"blink-liveview-websocket/common"
	"context"
	"io"
	"log"
	"net/http"
	"os/exec"
	"slices"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type CommandMessage struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{
	// TODO: Check if this is useful
	// EnableCompression: true,
}

var VALID_COMMANDS = []string{
	"liveview:start",
	"liveview:stop",
}

// The buffer size before dispatching the data to the WebSocket connection in bytes
var BUFFER_SIZE = 8 * 1024

// The idle timeout before closing the connection
var IDLE_TIMEOUT = 10 * time.Second

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
		"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
		// "-b:v", "1M",
		"-c:a", "aac", "-b:a", "128k",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-min_frag_duration", "500000", // 500ms fragments
		"-fflags", "nobuffer",
		"-flush_packets", "1", // Ensure FFmpeg writes data immediately
		"-f", "mp4", "pipe:1", // Output to stdout
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
	c.WriteJSON(CommandMessage{
		Command: "liveview:start",
		Data: map[string]interface{}{
			"message": "Liveview started",
		},
	})

	// Forward messages from the pipe to the WebSocket connection
	go func() {
		buf := make([]byte, BUFFER_SIZE)

		// Read from the pipe and write to the WebSocket connection
		for {
			n, err := outputPipe.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from pipe: %v", err)
				}
				break
			}

			if err := c.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("Error writing to WebSocket connection: %v", err)
				break
			}

			// Flush the buffer
			buf = make([]byte, BUFFER_SIZE)
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Wait for the ffmpeg command to finish
	if err := ffmpegCmd.Wait(); err != nil {
		log.Println("error waiting for ffplay", err)
	}

	// Tell the client that the liveview has stopped
	c.WriteJSON(CommandMessage{
		Command: "liveview:stop",
		Data: map[string]interface{}{
			"message": "Liveview stopped. Context cancelled",
		},
	})
}

// WebsocketHandler handles WebSocket connections from clients and performs upgrades
//
// w is the http.ResponseWriter
//
// r is the http.Request
//
// Example: http.HandleFunc("/ws", handlers.WebsocketHandler)
func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error upgrading the client", err)
		return
	}
	defer c.Close()

	var ctx context.Context
	var cancelCtx context.CancelFunc
	var lastMessage time.Time = time.Now()
	var liveviewStarted bool = false
	var closedClient bool = false

	// Monitor for idle connections
	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Check if the client has closed the connection
			if closedClient {
				return
			}

			// Check if the client has sent a message or if liveview has started
			if !liveviewStarted && time.Since(lastMessage) > IDLE_TIMEOUT {
				log.Println("Idle timeout reached. Closing connection")
				c.Close()
				return
			}
		}
	}()

	// Handle WebSocket IO
	for {
		var message CommandMessage
		if err := c.ReadJSON(&message); err != nil {
			break
		}

		if !slices.Contains(VALID_COMMANDS, message.Command) {
			log.Println("Invalid command received from client")
			break
		}

		lastMessage = time.Now()
		if message.Command == "liveview:start" {
			log.Println("Client requested liveview:start")

			ctx, cancelCtx = context.WithCancel(context.Background())
			go liveviewHandler(ctx, c, message.Data)
			liveviewStarted = true
		} else if message.Command == "liveview:stop" && liveviewStarted {
			log.Println("Client requested liveview:stop")
			cancelCtx()
			liveviewStarted = false
		}
	}

	closedClient = true

	if cancelCtx != nil {
		cancelCtx()
	}
}

// SetCheckOrigin sets the function to check the origin of the WebSocket connection
// This is useful for allowing connections from specific origins
//
// Example usage: handlers.SetCheckOrigin(func(r *http.Request) bool { return true })
func SetCheckOrigin(f func(r *http.Request) bool) {
	upgrader.CheckOrigin = f
}
