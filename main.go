package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const defaultPort int = 11111

func main() {
	port := flag.Int("port", defaultPort, "server port (default: 11111)")
	flag.Parse()

	// Create and start server
	server := NewServer(*port)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Handle graceful shutdown on Ctrl+C (SIGINT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // Wait for Ctrl+C
	log.Println("🛑 Received shutdown signal!")

	server.Shutdown() // Clean up rooms and close connections

	log.Println("👋 Server exited gracefully.")
}
