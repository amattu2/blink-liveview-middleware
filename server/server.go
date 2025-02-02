package server

import (
	"blink-liveview-websocket/common"
	"context"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type MessageData struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

var VALID_COMMANDS = []string{
	"liveview:start",
	"liveview:stop",
}

func liveviewHandler(ctx context.Context, c *websocket.Conn, data map[string]interface{}) {
	select {
	case <-ctx.Done():
		log.Println("Context cancelled")
		return
	default:
		region := data["account_region"].(string)
		token := data["api_token"].(string)
		account_id, _ := strconv.Atoi(data["account_id"].(string))
		network_id, _ := strconv.Atoi(data["network_id"].(string))
		camera_id, _ := strconv.Atoi(data["camera_id"].(string))
		device_type := data["camera_type"].(string)

		// TODO: Pipe the output back to the client as binary data
		// TODO: Kill the process when the client sends a stop command
		common.Livestream(region, token, device_type, account_id, network_id, camera_id)

		// Tell the client that the liveview has started
		c.WriteJSON(MessageData{
			Command: "liveview:start",
			Data: map[string]interface{}{
				"message": "Liveview started",
			},
		})
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
