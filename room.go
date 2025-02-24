package main

import (
	"fmt"
	"log"
)

const maxClients = 10

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
}

// NewRoom creates a new chat room instance with the given name.
// The room is initialized with all necessary channels and a random color.
// Returns a pointer to the newly created Room instance.
func NewRoom(name string) *Room {
	room := &Room{
		name:    name,
		forward: make(chan []byte),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]struct{}),
		quit:    make(chan struct{}),
		color:   getRandomColor(),
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
			client.send <- []byte(client.prompt)

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
		case msg := <-r.forward:
			for client := range r.clients {
				select {
				case client.send <- msg: // send the message
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
