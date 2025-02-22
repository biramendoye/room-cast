package main

import (
	"fmt"
	"time"
)

// Message represents a chat message exchanged over TCP.
type Message struct {
	Content   string
	Sender    string
	Timestamp time.Time
}

// NewMessage creates a new Message instance.
func NewMessage(content, sender string) Message {
	return Message{
		Content:   content,
		Sender:    sender,
		Timestamp: time.Now(),
	}
}

// Format returns a formatted string with a colored sender.
func (m Message) Format() string {
	return fmt.Sprintf("\n⏳ %s[%s]%s 🤖 %s%s%s 💬 %s\n",
		ColorDate, m.Timestamp.Format("2006-01-02 15:04:05"), ColorReset,
		ColorUser, m.Sender, ColorReset,
		m.Content,
	)
}
