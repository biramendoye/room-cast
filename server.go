package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type Server struct {
	// listener is the TCP listener used to accept incoming connections.
	listener net.Listener

	// port is the network port on which the server listens for connections.
	port int

	// clients stores all active client connections in a map for efficient management.
	clients map[*Client]struct{}

	// mu is a mutex used to synchronize access to shared resources like clients map.
	mu sync.RWMutex

	// stopChan receives signals indicating when the server should shut down gracefully.
	stopChan chan os.Signal

	// ctx provides a context that can be cancelled when shutting down the server,
	// allowing graceful termination of ongoing operations.
	ctx context.Context

	// cancel function cancels the context when needed (e.g., during shutdown).
	cancel context.CancelFunc
}

func NewServer(port int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:     port,
		clients:  make(map[*Client]struct{}),
		mu:       sync.RWMutex{},
		stopChan: make(chan os.Signal, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (srv *Server) run() {
	var err error

	srv.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", srv.port))
	if err != nil {
		log.Fatalf("Failed to start server on port %d: %v", srv.port, err)
	}
	defer srv.listener.Close()

	log.Println("Server is started listening on port:", srv.port)

	signal.Notify(srv.stopChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-srv.stopChan

		log.Println("Shutdown signal received. Stopping server...")

		srv.cancel()  // Cancel the context (stops `run()` loop)
		srv.Cleanup() // Cleanup all clients and close the listener
	}()

	// Accept incoming client connections
	for {
		select {
		case <-srv.ctx.Done(): // Stop accepting new connections
			log.Println("Server shutting down gracefully.")
			return
		default:
			conn, err := srv.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					log.Println("Server listener closed. Stopping...")
					return
				}
				log.Println("Failed to accept connection:", err)
				continue
			}
			go srv.handleConnection(conn)
		}
	}
}

func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	client := NewClient(conn)

	if err := client.setup(); err != nil {
		log.Println(err)
		return
	}

	srv.addClient(client)

	srv.handleClientMessages(client)
}

// handleClientMessages reads messages from the client and processes them.
func (srv *Server) handleClientMessages(client *Client) {
	buffer := make([]byte, 1024)

	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("Client %s disconnected", client.username)
			} else {
				log.Printf("Client %s read error: %v", client.username, err)
			}
			srv.removeClient(client)
			return
		}

		// Extract the actual message content
		rawMessage := strings.TrimSpace(string(buffer[:n]))
		if rawMessage == "" {
			client.sendPrompt()
			continue
		}

		message := NewMessage(rawMessage, client.username)

		// Broadcast the message to other clients
		srv.broadcastMessage(message, client)
	}
}

func (srv *Server) addClient(client *Client) {
	srv.mu.Lock()
	srv.clients[client] = struct{}{}
	srv.mu.Unlock()

	log.Printf("Client %s connected from IP: %s", client.username, client.conn.RemoteAddr().String())
}

func (srv *Server) removeClient(client *Client) {
	client.close()

	srv.mu.Lock()
	delete(srv.clients, client)
	srv.mu.Unlock()

	log.Printf("Client %s removed", client.username)
}

func (srv *Server) broadcastMessage(message Message, sender *Client) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	for client := range srv.clients {
		if client == sender {
			client.sendPrompt()
			continue
		}

		if err := client.sendMessage(message); err != nil {
			log.Printf("Error broadcasting to %s: %v", client.username, err)
			srv.removeClient(client)
		}
		client.sendPrompt()
	}
}

// Cleanup closes all connections and clears the clients map.
func (srv *Server) Cleanup() {
	log.Println("Cleaning up server...")

	// Unlock before closing client connections
	srv.mu.Lock()
	clients := srv.clients
	srv.clients = make(map[*Client]struct{}) // Reset clients map
	srv.mu.Unlock()

	for client := range clients {
		log.Printf("Closing connection for client %s", client.username)
		if closeErr := client.close(); closeErr != nil {
			log.Printf("Error closing connection for %s: %v", client.username, closeErr)
		}
	}

	// Close the listener safely
	if srv.listener != nil {
		srv.listener.Close()
	}

	log.Println("Server shutdown complete.")
}
