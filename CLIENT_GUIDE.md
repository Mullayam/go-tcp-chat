# TCP Chat Client Usage Guide

## New Layout: Server vs Client

The chat client now uses a clear visual distinction between server messages and your input:

- **Server Messages**: Prefixed with `S ->` (in Yellow/Red)
- **Client Input**: Prefixed with `C ->` (in Green)

Example:
```
S -> Welcome to TCP Chat Server!
S -> Please authenticate to continue.

S -> Enter your email address below
C -> user@example.com

S -> Please wait while we verify...
S -> Enter OTP code:
C -> 123456
```

## Available Clients

### 1. Standard Client (`chat-client.exe`)
**Features:**
- ✅ Clear `S ->` / `C ->` layout
- ✅ ANSI color support
- ✅ Robust input handling (ignores accidental empty lines)
- ✅ Color-coded message types:
  - **Yellow**: System/Server messages
  - **Red**: Errors
  - **Green**: Your input prompt
  - **Colored usernames**: Consistent per user

**Usage:**
```bash
go run cmd/client/main.go
# or
./bin/chat-client.exe
```

### 2. TUI Client (`chat-client-tui.exe`)
**Features:**
- ✅ Enhanced visual design with box drawing
- ✅ Visual indicators (●, ✗, 💬)
- ✅ Better message formatting

**Usage:**
```bash
go run cmd/client-tui/main.go
# or
./bin/chat-client-tui.exe
```

## Authentication Flow

1. **Email**: Enter your email when prompted.
2. **OTP**: Check your email for the code and enter it.
3. **Username**: Choose a unique username.

> **Note**: If you accidentally hit Enter multiple times, the server will now safely ignore the empty lines instead of failing authentication!

## Chat Experience

### Sending Messages
```
C -> hello everyone!
[alice]: hello everyone!
```

### Using Commands
```
C -> /users
S -> Online Users (3):
  - alice (you)
  - bob
```

## Building

```bash
# Build standard client
go build -o bin/chat-client.exe cmd/client/main.go
```
