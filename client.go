package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Client represents a chat client
type Client struct {
	// The name of the client
	username string

	// A TCP connection
	conn net.Conn

	// Visual color on terminal
	color string

	// Prompt format for the client
	prompt string
}

// NewClient initializes a new Client instance.
func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:  conn,
		color: getRandomColor(),
	}
}
func (c *Client) sendWelcomeMessage() error {
	logo := `
	▒█▀▀█ ▒█▀▀▀█ ▒█▀▀▀█ ▒█▀▄▀█ 　 ▒█▀▀█ ░█▀▀█ ▒█▀▀▀█ ▀▀█▀▀ 
	▒█▄▄▀ ▒█░░▒█ ▒█░░▒█ ▒█▒█▒█ 　 ▒█░░░ ▒█▄▄█ ░▀▀▀▄▄ ░▒█░░ 
	▒█░▒█ ▒█▄▄▄█ ▒█▄▄▄█ ▒█░░▒█ 　 ▒█▄▄█ ▒█░▒█ ▒█▄▄▄█ ░▒█░░ 
 `
	_, err := c.conn.Write([]byte(logo))
	if err != nil {
		return err
	}

	welcomeLines := []string{
		"🚀 Get ready for an awesome chat experience! 🚀\n",
		"		💡 Broadcasting Live: Join the fun! 🎤 💡\n",
		"				👉 To get started, please enter your username. 👈\n",
	}

	for _, line := range welcomeLines {
		_, err = c.conn.Write([]byte(line))
		if err != nil {
			return err
		}
		time.Sleep(time.Second) // Adds a cool typing effect
	}

	return err
}

// sendMessage sends a message to the client.
func (c *Client) sendMessage(msg Message) error {
	_, err := c.conn.Write([]byte(msg.Format()))
	return err
}

// sendPrompt sends to the client their prompt message
func (c *Client) sendPrompt() error {
	_, err := c.conn.Write([]byte(c.prompt))
	return err
}

func (c *Client) requestUsername() error {
	reader := bufio.NewReader(c.conn)

	for {
		username, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("Client disconnected before entering username")
			}
			return err
		}

		username = strings.TrimSpace(username)
		if len(username) >= 5 && len(username) <= 8 {
			c.username = username
			return nil
		}

		if _, err := c.conn.Write([]byte(fmt.Sprintf("%s 🚫 Invalid username! Please choose a name between 5 to 8 characters. Try again: %s\n", ColorRed, ColorReset))); err != nil {
			return err
		}
	}
}

// close safely closes the client's connection.
func (c *Client) close() error {
	return c.conn.Close()
}

// setup handles the client's initial setup process
func (c *Client) setup() error {
	if err := c.sendWelcomeMessage(); err != nil {
		return fmt.Errorf("error sending welcome message: %w", err)
	}

	if err := c.requestUsername(); err != nil {
		return fmt.Errorf("Error setting username: %w", err)
	}

	c.prompt = fmt.Sprintf("[%s%s%s] > ", c.color, c.username, ColorReset)

	// Ensure prompt is displayed for `netcat` users
	if err := c.sendPrompt(); err != nil {
		return fmt.Errorf("Error sending prompt: %w", err)
	}

	return nil
}
