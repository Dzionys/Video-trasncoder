package sse

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Broker struct {
	Clients        map[chan string]bool
	NewClients     chan chan string
	DefunctClients chan chan string
	Messages       chan string
}

var B *Broker

func (b *Broker) Start() {
	go func() {
		for {
			select {
			case s := <-b.NewClients:
				// start sending client messages
				b.Clients[s] = true
				log.Println("Added new client")

			case s := <-b.DefunctClients:
				// stop sending client messages
				delete(b.Clients, s)
				close(s)
				log.Println("Removed client")

			case msg := <-b.Messages:
				// there is a new message to send to all clients
				for s := range b.Clients {
					s <- msg
				}
				log.Printf("Broadcast message to %d clients", len(b.Clients))
			}
		}
	}()
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Channel for broker to send messages
	messageChan := make(chan string)

	// Add this client to the map of those that should receive updates
	b.NewClients <- messageChan

	// Listen to the closing of the http connection
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify

		// Remove this client from the map of attached clients when `EventHandler` exits
		b.DefunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	//w.Header().Set("Transfer-Encoding", "chunked")

	// Don't close the connection, instead loop endlessly
	for {
		msg, open := <-messageChan

		if !open {
			break
		}

		fmt.Fprintf(w, "data: %s\n\n", msg)

		// Flush data as soon as it has been updated
		f.Flush()
	}

	log.Println("Finished HTTP request at ", r.URL.Path)
}

func UpdateMessage(msg string) {
	B.Messages <- msg
}

func UpdateLogMessage(msg string) {
	curentTime := time.Now().Format("15:04:05")
	B.Messages <- fmt.Sprintf("<%v> %v", curentTime, msg)
}
