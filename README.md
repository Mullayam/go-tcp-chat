# TCP Chat Server

A production-ready TCP-based real-time chat server written in Go with email-based OTP authentication.

## Features

- ✅ **TCP Connection Management** - Accept and manage multiple client connections
- ✅ **Email OTP Authentication** - Secure authentication using one-time passwords
- ✅ **IP-Based Session Restrictions** - One active session per IP address
- ✅ **Real-Time Messaging** - Instant message delivery
- ✅ **Room Management** - Public and private chat rooms
- ✅ **Private Messaging** - Direct 1-to-1 conversations
- ✅ **In-Memory Storage** - No persistent data storage
- ✅ **Graceful Cleanup** - Automatic session cleanup on disconnect

## Prerequisites

- Go 1.21 or higher
- SMTP server credentials (Gmail, SendGrid, Mailgun, etc.)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/mullayam/go-tcp-chat.git
cd go-tcp-chat
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your SMTP credentials
```

## Configuration

Create a `.env` file in the project root:

```env
TCP_PORT=8888

# SMTP Configuration for Email OTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-specific-password

# OTP Settings
OTP_EXPIRATION_MINUTES=5
OTP_MAX_RETRIES=3

# Username Validation
USERNAME_MIN_LENGTH=3
USERNAME_MAX_LENGTH=16
```

### Gmail Setup

To use Gmail for sending OTP emails:

1. Enable 2-Factor Authentication on your Google account
2. Generate an App Password:
   - Go to Google Account Settings → Security
   - Under "2-Step Verification", select "App passwords"
   - Generate a new app password for "Mail"
   - Use this password in `SMTP_PASSWORD`

## Running the Server

```bash
go run cmd/server/main.go
```

Or build and run:

```bash
go build -o chat-server cmd/server/main.go
./chat-server
```

## Connecting to the Server

You can connect using any TCP client. Examples:

### Using Telnet

```bash
telnet localhost 8888
```

### Using Netcat

```bash
nc localhost 8888
```

### Using a Custom Go Client

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8888")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Read from server
    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }()

    // Write to server
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        fmt.Fprintf(conn, "%s\n", scanner.Text())
    }
}
```

## Authentication Flow

1. Connect to the server
2. Enter your email address
3. Check your email for the 6-digit OTP code
4. Enter the OTP code
5. Choose a username (3-16 characters, alphanumeric + underscore)
6. Start chatting!

## Available Commands

Once authenticated, you can use the following commands:

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/users` | List all online users |
| `/rooms` | List all available rooms |
| `/join <room>` | Join or create a room (e.g., `/join #gaming`) |
| `/leave` | Leave current room and return to #general |
| `/msg <user> <message>` | Send a private message to a user |
| `/quit` | Disconnect from the server |

## Usage Examples

### Joining a Room

```
/join #gaming
*** You joined #gaming ***
*** alice joined the room ***
```

### Sending Messages

```
Hello everyone!
[bob]: Hello everyone!
```

### Private Messaging

```
/msg alice Hey, how are you?
[PM to alice]: Hey, how are you?
```

### Listing Online Users

```
/users
Online Users (3):
  - alice
  - bob (you)
  - charlie
```

## Architecture

```
go-tcp-chat/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── server/
│   │   └── tcp_server.go        # TCP server implementation
│   ├── session/
│   │   ├── manager.go           # Session and IP management
│   │   └── session.go           # Session model
│   ├── auth/
│   │   ├── otp.go               # OTP generation and validation
│   │   └── email.go             # Email service
│   ├── room/
│   │   ├── manager.go           # Room management
│   │   └── room.go              # Room model
│   ├── message/
│   │   ├── router.go            # Message routing
│   │   └── handler.go           # Command handling
│   └── protocol/
│       └── protocol.go          # Protocol definitions
└── config/
    └── config.go                # Configuration
```

## Security Features

- **No Persistent Storage** - All data exists only in memory
- **IP-Based Restrictions** - One connection per IP address
- **OTP Expiration** - OTPs expire after 5 minutes (configurable)
- **One-Time Use** - OTPs can only be used once
- **Max Retry Limits** - Prevents brute force attacks
- **Email Validation** - Validates email format before sending OTP
- **Username Validation** - Enforces username rules and uniqueness

## Limitations

- **In-Memory Only** - All data is lost on server restart
- **No Message History** - Messages are not stored or logged
- **Single Server** - Not designed for horizontal scaling
- **No Encryption** - TCP connections are not encrypted (use SSH tunneling for production)

## Production Deployment

For production use, consider:

1. **TLS/SSL** - Wrap TCP connections in TLS
2. **Reverse Proxy** - Use nginx or similar for connection management
3. **Rate Limiting** - Implement rate limiting for OTP requests
4. **Monitoring** - Add metrics and logging
5. **Load Balancing** - Use sticky sessions if scaling horizontally

## Troubleshooting

### "Failed to send OTP" Error

- Check SMTP credentials in `.env`
- Verify SMTP server allows connections from your IP
- For Gmail, ensure you're using an App Password, not your regular password

### "IP address already has an active connection"

- Only one connection per IP is allowed
- Disconnect the existing connection first
- Wait a few seconds for cleanup to complete

### "Username already taken"

- Choose a different username
- Usernames are unique across all active sessions

## License

MIT License - feel free to use this project for learning or production use.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

Created by Mullayam
