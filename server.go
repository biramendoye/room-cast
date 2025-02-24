package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	// listener is the TCP listener used to accept incoming connections.
	listener net.Listener

	// port is the network port on which the server listens for connections.
	port int

	// clients stores all active client connections in a map for efficient management.
	rooms map[string]*Room

	// mu is a mutex used to synchronize access to shared resources like rooms map.
	mu sync.RWMutex
}

func NewServer(port int) *Server {
	return &Server{
		rooms: make(map[string]*Room),
		port:  port,
	}
}

// Start begins listening for client connections.
func (srv *Server) Start() error {
	addr := fmt.Sprintf(":%d", srv.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	srv.listener = ln
	log.Println("✅ Server started on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// log.Println("Accept error:", err)
			continue
		}

		go srv.handleConnection(conn)
	}

}

// handleConnection manages a new client connection.
func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// Get valid username and room name
	username, roomName, err := s.setupClient(conn, reader)
	if err != nil {
		log.Printf("🚨 Failed to setup client: %v\n", err)
		return
	}

	// -------------------------------------------
	room := s.getOrCreateRoom(roomName)

	client := NewClient(conn, username, room)

	room.join <- client

	go client.read()
	go client.write()
}

// getOrCreateRoom finds an existing room or creates a new one.
func (s *Server) getOrCreateRoom(name string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exist := s.rooms[name]; exist {
		return room
	}

	// Create a new room if no available space
	newRoom := NewRoom(name)

	s.rooms[name] = newRoom
	log.Printf("🏠 Room %s created.\n", name)
	go newRoom.run()
	return newRoom
}

// Shutdown gracefully shuts down the server.
func (srv *Server) Shutdown() {
	log.Println("⚠️  Shutting down server...")
	if srv.listener != nil {
		srv.listener.Close()
	}

	// Close all rooms
	srv.mu.Lock()
	for name, room := range srv.rooms {
		room.stop()
		delete(srv.rooms, name)
	}
	srv.mu.Unlock()

	log.Println("✅ Server shut down gracefully.")
}

// setupClient prompts the user until a valid username and room name are entered.
func (s *Server) setupClient(conn net.Conn, reader *bufio.Reader) (string, string, error) {
	// send Welcome Message
	if err := sendWelcomeMessage(conn); err != nil {
		log.Printf("🚨 Failed to send welcome message: %v\n", err)
		return "", "", err
	}

	var username, roomName string

	// Keep asking for username until it's valid
	for {
		conn.Write([]byte("Enter username: "))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", "", fmt.Errorf("error reading username: %w", err)
		}
		username = strings.TrimSpace(input)

		if isValidUsername(username) {
			break
		}
		conn.Write([]byte("❌ Invalid username. Must be 3-15 characters (A-Z, a-z, 0-9, _).\n"))
	}

	// Keep asking for room name until it's valid
	for {
		conn.Write([]byte("Enter room name: "))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", "", fmt.Errorf("error reading room name: %w", err)
		}
		roomName = strings.TrimSpace(input)

		if isValidRoomName(roomName) {
			break
		}
		conn.Write([]byte("❌ Invalid room name. Must be 3-20 characters (A-Z, a-z, 0-9, _).\n"))
	}

	return strings.ToLower(username), strings.ToUpper(roomName), nil
}

func sendWelcomeMessage(conn net.Conn) error {
	logo := `
	▒█▀▀█ ▒█▀▀▀█ ▒█▀▀▀█ ▒█▀▄▀█ 　 ▒█▀▀█ ░█▀▀█ ▒█▀▀▀█ ▀▀█▀▀ 
	▒█▄▄▀ ▒█░░▒█ ▒█░░▒█ ▒█▒█▒█ 　 ▒█░░░ ▒█▄▄█ ░▀▀▀▄▄ ░▒█░░ 
	▒█░▒█ ▒█▄▄▄█ ▒█▄▄▄█ ▒█░░▒█ 　 ▒█▄▄█ ▒█░▒█ ▒█▄▄▄█ ░▒█░░ 
 `
	_, err := conn.Write([]byte(logo))
	if err != nil {
		return err
	}

	welcomeLines := []string{
		"🚀 Get ready for an awesome chat experience! 🚀\n",
		"💡 Broadcasting Live: Join the fun! 🎤 💡\n",
		"👉 To get started, please enter your username. 👈\n",
	}

	for _, line := range welcomeLines {
		_, err = conn.Write([]byte(line))
		if err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond) // Adds a cool typing effect
	}

	return err
}
