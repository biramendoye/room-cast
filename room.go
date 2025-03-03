package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const maxClients = 10
const maxHistory = 100

// Room represents a chat room where clients can communicate.
// The room name is displayed in a unique color in the terminal for visual distinction.
type Room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan []byte

	// join is a channel for clients wishing to join the room.
	join chan *Client

	// leave is a channel for clients wishing to leave the room.
	leave chan *Client

	// quit is a channel used to signal the room to shut down
	quit chan struct{}

	// clients holds all current clients in this room.
	clients map[*Client]struct{}

	// name is the identifier of the room, displayed in the terminal
	// with a unique color for visual distinction.
	name string

	// color is the ANSI color code used to display the room name
	// in the terminal. Each room has a unique color for better
	// visual organization.
	color string

	historyFile string

	mu sync.Mutex
}

// NewRoom creates a new chat room instance with the given name.
// The room is initialized with all necessary channels and a random color.
// Returns a pointer to the newly created Room instance.
func NewRoom(name string) *Room {
	room := &Room{
		name:        name,
		forward:     make(chan []byte),
		join:        make(chan *Client),
		leave:       make(chan *Client),
		clients:     make(map[*Client]struct{}),
		quit:        make(chan struct{}),
		color:       getRandomColor(),
		historyFile: fmt.Sprintf("history_%s", name),
	}

	return room
}

// run manages the core chat room operations including:
// - Handling client joins and leaves.
// - Forwarding messages to all connected clients
// - Managing room capacity
// - Handling shutdown signals
// This method runs in an infinite loop until the room is stopped.
func (r *Room) run() {
	log.Printf("🔄 Starting room %s\n", r.name)

	for {
		select {
		// joining
		case client := <-r.join:
			if len(r.clients) >= maxClients {
				log.Printf("❌ Room %s is full. %s cannot join.\n", r.name, client.username)
				errorMessage := fmt.Sprintf("❌ Room %s is full. Cannot join.\n", r.name)
				client.writeMessage([]byte(errorMessage))

				client.close()
				continue
			}
			r.clients[client] = struct{}{}
			log.Printf("✅ %s joined %s", client.username, r.name)
			// client.send <- []byte(client.prompt)
			client.writeMessage([]byte(client.prompt))

			// Notify others
			r.broadcast(&Message{
				Content: fmt.Sprintf("📢 %s has joined the room.\n", client.username),
				Sender:  client.username,
				Type:    NotificationType,
			})

		// leaving
		case client := <-r.leave:
			r.removeClient(client)

			// Notify others
			r.broadcast(&Message{
				Content: fmt.Sprintf("📢 %s has left the room.\n", client.username),
				Sender:  client.username,
				Type:    NotificationType,
			})

		// forward message to all clients
		case msgBytes := <-r.forward:
			var msg Message
			err := json.Unmarshal(msgBytes, &msg)
			if err != nil {
				log.Printf("❌ Failed to parse message JSON: %v", err)
				continue
			}
			r.saveMessageToFile(msg)

			for client := range r.clients {
				select {
				case client.send <- msgBytes: // send the message
				default:
					// failed to send
					log.Printf("❌ Failed to send message to %s in room %s", client.username, r.name)
					r.removeClient(client)
				}
			}

		case <-r.quit:
			log.Printf("🛑 Shutting down room %s", r.name)
			for client := range r.clients {
				close(client.send)
				delete(r.clients, client)
			}
			log.Printf("✅ Room %s shutdown complete", r.name)
			return
		}
	}
}

// stop gracefully shuts down the room by:
// - Closing all control channels
// - Removing all connected clients
// - Resetting the room's client map
func (r *Room) stop() {
	close(r.quit)
}

func (r *Room) removeClient(client *Client) {
	if _, exists := r.clients[client]; exists {
		delete(r.clients, client)
		close(client.send)

		if client.conn != nil {
			client.conn.Close()
			client.conn = nil
		}

		log.Printf("✅ %s left %s", client.username, r.name)
	}
}

func (r *Room) broadcast(msg *Message) {
	jsonMessage := msg.ToJSON()
	for client := range r.clients {
		if client.username != msg.Sender { // Exclude the sender
			client.send <- jsonMessage
		}
	}
}

func (r *Room) saveMessageToFile(msg Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ Error saving message for %s: %v", r.name, err)
		return
	}
	defer file.Close()

	_, err = file.Write(bytes.TrimPrefix(msg.formatAndConvertToBytes(), []byte("\n")))
	if err != nil {
		log.Printf("❌ Error writing message to file: %v", err)
	}
}

// sendHistory reads the entire history file and sends it to the client
func (r *Room) sendHistory(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.Open(r.historyFile)
	if err != nil {
		client.writeMessage([]byte("📭 No chat history available.\n"))
		return
	}
	defer file.Close()

	// Read entire file content
	historyBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("❌ Error reading history file: %v", err)
		client.writeMessage([]byte("❌ Failed to load chat history.\n"))
		return
	}

	client.writeMessage([]byte("📜 Previous messages:\n"))
	client.writeMessage(historyBytes)
}
