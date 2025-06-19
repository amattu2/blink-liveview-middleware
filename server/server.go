package server

import (
	"blink-liveview-websocket/handlers"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"time"
)

func Run(address string, env string, origins []string, classificationEnabled bool, classificationInterval int) {
	server := &http.Server{Addr: address}

	http.HandleFunc("/liveview", handlers.WebsocketHandler)

	if env == "development" {
		log.Println("Enabled static file server")
		http.Handle("/", http.FileServer(http.Dir("./static")))
	}

	if len(origins) > 0 {
		log.Println("Enabled custom WebSocket origins", origins)
		handlers.SetCheckOrigin(func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			if origins[0] == "*" {
				return true
			}

			return slices.Contains(origins, origin)
		})
	}

	if classificationEnabled {
		log.Println("Enabled classification on interval", classificationInterval, "seconds")
		handlers.SetClassificationEnabled(true)
		handlers.SetClassificationInterval(time.Duration(classificationInterval) * time.Second)
	} else {
		handlers.SetClassificationEnabled(false)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("Received SIGINT. Shutting down...")

		ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCtx()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("HTTP shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}

	log.Println("Ignoring new requests. Waiting for existing requests to finish...")
}
