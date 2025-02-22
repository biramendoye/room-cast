package main

import "flag"

const defaultPort int = 11111

func main() {
	port := flag.Int("port", defaultPort, "server port (default: 11111)")
	flag.Parse()

	// Create and start server
	server := NewServer(*port)
	server.run()
}
