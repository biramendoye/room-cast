package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

const messageBufferSize = 256

// Client represents a single chatting user
type Client struct {
	// The name of the client
	username string

	// conn is the TCP connection for this client.
	conn net.Conn

	// send is a channel on which messages are sent.
	send chan []byte

	// room is the room this client is chatting in.
	room *Room

	// Prompt format for the client
	prompt string
}

func NewClient(conn net.Conn, username string, room *Room) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan []byte, messageBufferSize),
		room:     room,
		username: username,
		prompt:   fmt.Sprintf("%s%s 🏠 %s%s > ", room.color, username, room.name, ColorReset),
	}
}

// read allows our client to read from the TCP conn,
// continually sending any received messages to the
// forward channel on the room type.
// If it encounters an error, the loop will break and the conn will be closed.
func (c *Client) read() {
	reader := bufio.NewReader(c.conn)
	for {
		if _, err := c.conn.Write([]byte(c.prompt)); err != nil {
			log.Printf("🚨Error writing prompt: %v", err)
			break
		}

		msg, err := reader.ReadBytes('\n')
		if err != nil {
			log.Printf("🚨Read error: %v", err)
			break
		}

		msg = bytes.TrimSpace(msg)
		if len(msg) == 0 {
			continue
		}

		message := Message{
			Content:   string(msg),
			Sender:    c.username,
			Timestamp: time.Now(),
		}

		c.room.forward <- message.ToJSON()
	}

	c.close()
}

// write continually accepts messages from the send channel,
// it write everything out of the conn.
// If writing to the conn fails, the for loop is broken and the conn is closed.
func (c *Client) write() {
	for rawMessage := range c.send {
		msg, err := FromJSON(rawMessage)
		if err != nil {
			log.Printf("❌ Failed to parse message: %v", err)
			continue
		}
		if msg.Sender == c.username {
			continue
		}

		if err := c.writeMessage(msg.formatAndConvertToBytes()); err != nil {
			log.Printf("🚨Write error: %v", err)
			break
		}

		if _, err := c.conn.Write([]byte(c.prompt)); err != nil {
			log.Printf("🚨Error writing prompt: %v", err)
		}
	}
	c.close()
}

// writeMessage writes the message to the TCP connection of the client
func (c *Client) writeMessage(msg []byte) error {
	_, err := c.conn.Write(msg)
	return err
}

func (c *Client) close() {
	// Ensure cleanup only happens once
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Notify the room that this client is leaving
	if c.room != nil {
		c.room.leave <- c
	}
}
