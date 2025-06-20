package handlers

import (
	"blink-liveview-websocket/common"
	"context"
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
	"liveview:classify",
}

// The buffer size before dispatching the data to the WebSocket connection in bytes
var BUFFER_SIZE int = 4 * 1024

// The idle timeout before closing the connection
var IDLE_TIMEOUT time.Duration = 10 * time.Second

// A flag to enable or disable classification features
var CLASSIFICATION bool = false

func liveviewHandler(ctx context.Context, c *websocket.Conn, details map[string]any, buffer *common.CircularBuffer) {
	region := details["account_region"].(string)
	token := details["api_token"].(string)
	account_id, _ := strconv.Atoi(details["account_id"].(string))
	network_id, _ := strconv.Atoi(details["network_id"].(string))
	camera_id, _ := strconv.Atoi(details["camera_id"].(string))
	device_type := details["camera_type"].(string)

	// FFmpeg command to process raw video stream
	streamCmd := exec.Command("ffmpeg",
		"-i", "pipe:0", // Read from stdin
		"-c:v", "libx264", // Use H.264 codec for video
		"-preset", "ultrafast", // Use ultrafast preset for low latency
		"-tune", "zerolatency", // Tune for low latency
		"-g", "30", // Set GOP size to 30 frames
		"-keyint_min", "30", // Set minimum keyframe interval to 30 frames
		"-sc_threshold", "0", // Disable scene change triggering new GOPs
		"-c:a", "aac", "-b:a", "128k", // Use AAC codec for audio with 128k bitrate
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-min_frag_duration", "100000", // 100ms fragments
		"-flags", "low_delay", // Set low delay flags
		"-flush_packets", "1", // Flush packets immediately
		"-f", "mp4", // Output format MP4
		"pipe:1", // Output to stdout
	)

	streamIn, streamInErr := streamCmd.StdinPipe()
	streamOut, streamOutErr := streamCmd.StdoutPipe()
	if streamInErr != nil || streamOutErr != nil {
		log.Println("error creating ffmpeg stream pipes", streamInErr, streamOutErr)
		return
	}
	defer streamIn.Close()
	defer streamOut.Close()

	if err := streamCmd.Start(); err != nil {
		log.Println("error starting ffmpeg stream command", err)
		return
	}
	defer streamCmd.Process.Kill()

	// TODO: Handle common.Livestream errors and propagate them to the client
	// The client currently has no idea if the livestream internally failed
	go common.Livestream(ctx, common.AccountDetails{
		Region:     region,
		Token:      token,
		DeviceType: device_type,
		AccountId:  account_id,
		NetworkId:  network_id,
		CameraId:   camera_id,
	}, streamIn)

	// Communicate the start and stop of the liveview to the client
	c.WriteJSON(CommandMessage{
		Command: "liveview:start",
		Data: map[string]interface{}{
			"message": "Liveview started",
		},
	})
	defer c.WriteJSON(CommandMessage{
		Command: "liveview:stop",
		Data: map[string]interface{}{
			"message": "Liveview stopped",
		},
	})

	// Forward messages from the processed stream to the WebSocket connection
	go func() {
		buf := make([]byte, BUFFER_SIZE)

		for {
			n, err := streamOut.Read(buf)
			if err != nil {
				log.Printf("Error reading from ffmpeg stdout: %v", err)
				break
			}

			if err := c.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("Error writing to WebSocket connection: %v", err)
				break
			}

			if CLASSIFICATION {
				// Write the data to the circular buffer for classification
				buffer.Write(buf[:n])
			}

			// Flush the buffer
			buf = make([]byte, BUFFER_SIZE)
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Wait for the ffmpeg command to finish
	if err := streamCmd.Wait(); err != nil {
		log.Println("error waiting for ffmpeg stream command", err)
	}
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
	var buffer *common.CircularBuffer
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
			break
		}

		lastMessage = time.Now()
		if message.Command == "liveview:start" && !liveviewStarted {
			log.Println("Client requested liveview:start")

			ctx, cancelCtx = context.WithCancel(context.Background())
			buffer = common.NewCircularBuffer(100) // TODO: Evaluate buffer size
			go liveviewHandler(ctx, c, message.Data, buffer)
			liveviewStarted = true
		} else if message.Command == "liveview:stop" && liveviewStarted {
			log.Println("Client requested liveview:stop")
			cancelCtx()
			liveviewStarted = false
		} else if message.Command == "liveview:classify" && CLASSIFICATION && liveviewStarted {
			log.Println("Client requested liveview:classify")
			// TODO: Implement classification logic
			c.WriteJSON(CommandMessage{
				Command: "liveview:classify_response", // TODO: name this command properly
				Data: map[string]any{
					"message": "Classification response message",
					"labels":  []string{"label1", "label2"}, // Example labels
				},
			})
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

// SetClassificationEnabled sets the flag to enable or disable classification features
// This allows clients to request classification of liveview streams, but does not happen
// automatically unless the client requests it.
//
// Example usage: handlers.SetClassificationEnabled(true)
func SetClassificationEnabled(enabled bool) {
	CLASSIFICATION = enabled
}
