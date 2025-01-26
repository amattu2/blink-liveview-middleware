package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var address = flag.String("address", ":8080", "HTTP server address")
var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/liveview", echo)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Fatal(http.ListenAndServe(*address, nil))
}
