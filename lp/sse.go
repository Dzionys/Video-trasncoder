package lp

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Broker struct {
	Clients        map[chan string]string
	NewClients     chan Client
	DefunctClients chan chan string
	Messages       chan Message
}

type Client struct {
	ClientChannel chan string
	ClientId      string
}

type Message struct {
	Msg      string
	ClientId string
}

var (
	B       *Broker
	started = false
)

func (b *Broker) Start() {
	started = true
	go func() {
		for {
			select {
			case s := <-b.NewClients:
				// start sending client messages
				b.Clients[s.ClientChannel] = s.ClientId

			case s := <-b.DefunctClients:
				// stop sending client messages
				delete(b.Clients, s)
				close(s)

			case msg := <-b.Messages:
				// there is a new message to send to all clients
				for s, _ := range b.Clients {
					s <- msg.Msg
				}
				// log.Printf("Broadcast message to %d clients", len(b.Clients))
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
	client := Client{
		messageChan,
		r.RemoteAddr,
	}
	b.NewClients <- client

	// Listen to the closing of the http connection
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify

		// Remove this client from the map of attached clients when `EventHandler` exits
		b.DefunctClients <- messageChan
		//log.Println("HTTP connection just closed.")
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

	//log.Println("Finished HTTP request at ", r.URL.Path)
}

func UpdateMessage(msg string) {
	if !started {
		log.Println("SSE has not been server started")
		return
	}
	message := Message{
		msg,
		"",
	}
	B.Messages <- message
}

func UpdateLogMessage(msg string, clid string) {
	if !started {
		log.Println("SSE has not been server started")
		return
	}
	curentTime := time.Now().Format("15:04:05")
	message := Message{
		fmt.Sprintf("<%v> %v", curentTime, msg),
		clid,
	}
	B.Messages <- message
}
