package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	NotificationType = "Notification"
	UserMessageType  = "UserMessage"
)

// Message represents a chat message exchanged over TCP.
type Message struct {
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// NewMessage creates a new Message instance.
func NewMessage(content, sender, msgType string) Message {
	return Message{
		Content:   content,
		Sender:    sender,
		Timestamp: time.Now(),
		Type:      msgType,
	}
}

// formatAndConvertToBytes formats the message with colors and converts it to bytes.
// Returns the formatted message as JSON bytes and any error that occurred.
func (m Message) formatAndConvertToBytes() []byte {
	if m.Type == NotificationType {
		return []byte(fmt.Sprintf("\n%s%s%s", ColorNotification, m.Content, ColorReset))
	}

	formatted := fmt.Sprintf("\n⏳ %s[%s] 🤖 %s 💬 %s%s\n",
		ColorWhiteText, m.Timestamp.Format("2006-01-02 15:04:05"), m.Sender, m.Content, ColorReset,
	)

	// Convert to JSON bytes
	return []byte(formatted)
}

// ToJSON converts the message to a JSON byte array.
func (m Message) ToJSON() []byte {
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("❌ Error encoding message to JSON: %v\n", err)
		return nil
	}
	return jsonData
}

// FromJSON decodes the JSON byte array into a Message struct.
func FromJSON(jsonData []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(jsonData, &msg)
	if err != nil {
		return Message{}, fmt.Errorf("❌ Error decoding JSON: %v", err)
	}
	return msg, nil
}
