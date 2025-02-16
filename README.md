# room-cast

room-cast recreates the functionality of NetCat in a Server-Client Architecture. It allows clients to connect via TCP, join specific rooms, and broadcast messages within those rooms. Ideal for building room-based communication systems with efficient message handling in Go.

## 📋 Features 📋

- ✨ Real-time message broadcasting within rooms
- 🔒 Connection limit enforcement (max 10 clients per rooms)
- ⚡ Concurrent client handling
- 🔄 Automatic disconnection cleanup
- 💻 Thread-safe operations
- 🌐 TCP/IP protocol implementation
- 🖥️ NetCat (nc) compatible
- 📢 Room-based message isolation

## 📚 Technical Details 📚

- 🔧 Built with Go's net package for TCP communications
- 🔄 Uses goroutines for concurrent connection handling
- 🔒 Implements mutex synchronization for thread safety
- 💻 Leverages channels for message broadcasting
- ⚙️ Includes proper error handling and resource cleanup

## 👥 Architecture 👥

- 🏢 Centralized server managing all client connections
- 📱 Clients connect via the netcat command (`nc`)
- 🔄 Message broadcasting system within rooms
- 🔒 Connection limit enforcement mechanism

## 📂 Project Structure

```
room-cast/
│── cmd/
│ ├── main.go
│
│── internal/
│ ├── server/
│ │ ├── server.go
│ │ ├── handler.go
│ │ ├── client.go
│ │ ├── room.go
│ ├── protocol/
│ │ ├── message.go
│
│── Makefile
│── README.md
│── go.mod
│── go.sum
```

### Key Directories:

    - cmd/: Entry point for the server application.
    - internal/: Core logic including server, client, room management, and message protocol.
    - Makefile: Build and run commands.

## 💻 Getting Started 💻

1. Start the server

```bash
go run cmd/main.go --port <port-number>
```

2. connect to the server using nc command:

```bash
nc <server-address> 11111
```

## 📝 Usage Instructions 📝

- 📊 Server starts on port `11111` by default
- 🖥️ Clients automatically connect to (nc localhost 11111)
- 📝 Join a room by sending `/join <room-name>`
- ✍️ Type messages and press Enter to send
- 📤 Messages appear instantly on all connected clients in the same room
- ❌ Press Ctrl+C to exit cleanly
- 📝 Use `/leave` to leave current room
- 📝 Use `/rooms` to list available rooms

## 🎯 Learning Outcomes 🎯

- 📚 Understanding TCP/IP networking fundamentals
- 🔄 Mastering Go's concurrency features
- 🔒 Implementing thread-safe operations
- 💻 Building scalable network applications
- 📱 Creating real-time communication systems
- 📝 Implementing room-based message routing
