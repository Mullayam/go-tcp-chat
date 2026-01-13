package message

import (
	"fmt"
	"strings"

	"github.com/mullayam/go-tcp-chat/internal/protocol"
	"github.com/mullayam/go-tcp-chat/internal/room"
	"github.com/mullayam/go-tcp-chat/internal/session"
)

// Handler handles command processing
type Handler struct {
	sessionMgr *session.Manager
	roomMgr    *room.Manager
}

// NewHandler creates a new command handler
func NewHandler(sessionMgr *session.Manager, roomMgr *room.Manager) *Handler {
	return &Handler{
		sessionMgr: sessionMgr,
		roomMgr:    roomMgr,
	}
}

// HandleCommand processes a command from a user
func (h *Handler) HandleCommand(sess *session.Session, command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		return h.handleHelp(sess)
	case "/users":
		return h.handleUsers(sess)
	case "/rooms":
		return h.handleRooms(sess)
	case "/join":
		return h.handleJoin(sess, parts)
	case "/leave":
		return h.handleLeave(sess)
	case "/msg":
		return h.handlePrivateMessage(sess, parts)
	case "/quit":
		return h.handleQuit(sess)
	default:
		return sess.Send(protocol.NewErrorMessage(fmt.Sprintf("Unknown command: %s. Type /help for available commands.", cmd)).Format())
	}
}

// handleHelp shows available commands
func (h *Handler) handleHelp(sess *session.Session) error {
	help := `
Available Commands:
  /help              - Show this help message
  /users             - List all online users
  /rooms             - List all available rooms
  /join <room>       - Join or create a room
  /leave             - Leave current room and return to #general
  /msg <user> <msg>  - Send a private message to a user
  /quit              - Disconnect from the server

Chat:
  - Type any message to chat in your current room
  - Messages are only visible to users in the same room
`
	return sess.Send(protocol.NewCommandMessage(help).Format())
}

// handleUsers lists all online users
func (h *Handler) handleUsers(sess *session.Session) error {
	usernames := h.sessionMgr.GetOnlineUsernames()
	if len(usernames) == 0 {
		return sess.Send(protocol.NewCommandMessage("No users online.").Format())
	}

	msg := fmt.Sprintf("Online Users (%d):\n", len(usernames))
	for _, username := range usernames {
		if username == sess.GetUsername() {
			msg += fmt.Sprintf("  - %s (you)\n", username)
		} else {
			msg += fmt.Sprintf("  - %s\n", username)
		}
	}
	return sess.Send(protocol.NewCommandMessage(msg).Format())
}

// handleRooms lists all available rooms
func (h *Handler) handleRooms(sess *session.Session) error {
	roomNames := h.roomMgr.GetAllRoomNames()
	if len(roomNames) == 0 {
		return sess.Send(protocol.NewCommandMessage("No rooms available.").Format())
	}

	msg := fmt.Sprintf("Available Rooms (%d):\n", len(roomNames))
	for _, roomName := range roomNames {
		roomType, memberCount, _ := h.roomMgr.GetRoomInfo(roomName)
		currentRoom := sess.GetCurrentRoom()
		if roomName == currentRoom {
			msg += fmt.Sprintf("  - %s [%s] (%d members) (current)\n", roomName, roomType, memberCount)
		} else {
			msg += fmt.Sprintf("  - %s [%s] (%d members)\n", roomName, roomType, memberCount)
		}
	}
	return sess.Send(protocol.NewCommandMessage(msg).Format())
}

// handleJoin joins or creates a room
func (h *Handler) handleJoin(sess *session.Session, parts []string) error {
	if len(parts) < 2 {
		return sess.Send(protocol.NewErrorMessage("Usage: /join <room>").Format())
	}

	roomName := parts[1]
	if !strings.HasPrefix(roomName, "#") {
		roomName = "#" + roomName
	}

	// Leave current room
	currentRoom := sess.GetCurrentRoom()
	if currentRoom != "" {
		h.roomMgr.LeaveRoom(sess)
		if room, exists := h.roomMgr.GetRoom(currentRoom); exists {
			room.Broadcast(protocol.NewSystemMessage(fmt.Sprintf("%s left the room", sess.GetUsername())), "")
		}
	}

	// Create room if it doesn't exist
	room, err := h.roomMgr.CreateRoom(roomName)
	if err != nil {
		return sess.Send(protocol.NewErrorMessage(err.Error()).Format())
	}

	// Join the room
	err = h.roomMgr.JoinRoom(roomName, sess)
	if err != nil {
		return sess.Send(protocol.NewErrorMessage(err.Error()).Format())
	}

	// Notify user
	sess.Send(protocol.NewSystemMessage(fmt.Sprintf("You joined %s", roomName)).Format())

	// Notify room members
	room.Broadcast(protocol.NewSystemMessage(fmt.Sprintf("%s joined the room", sess.GetUsername())), sess.GetUsername())

	return nil
}

// handleLeave leaves the current room
func (h *Handler) handleLeave(sess *session.Session) error {
	currentRoom := sess.GetCurrentRoom()
	if currentRoom == protocol.DefaultRoom {
		return sess.Send(protocol.NewErrorMessage("You are already in the default room.").Format())
	}

	if currentRoom == "" {
		return sess.Send(protocol.NewErrorMessage("You are not in any room.").Format())
	}

	// Notify room before leaving
	if room, exists := h.roomMgr.GetRoom(currentRoom); exists {
		room.Broadcast(protocol.NewSystemMessage(fmt.Sprintf("%s left the room", sess.GetUsername())), "")
	}

	// Leave current room
	h.roomMgr.LeaveRoom(sess)

	// Join default room
	defaultRoom := h.roomMgr.GetDefaultRoom()
	defaultRoom.AddMember(sess)
	sess.SetCurrentRoom(protocol.DefaultRoom)

	// Notify user
	sess.Send(protocol.NewSystemMessage(fmt.Sprintf("You left %s and returned to %s", currentRoom, protocol.DefaultRoom)).Format())

	// Notify default room
	defaultRoom.Broadcast(protocol.NewSystemMessage(fmt.Sprintf("%s joined the room", sess.GetUsername())), sess.GetUsername())

	return nil
}

// handlePrivateMessage sends a private message
func (h *Handler) handlePrivateMessage(sess *session.Session, parts []string) error {
	if len(parts) < 3 {
		return sess.Send(protocol.NewErrorMessage("Usage: /msg <username> <message>").Format())
	}

	targetUsername := parts[1]
	message := strings.Join(parts[2:], " ")

	// Check if target user exists
	targetSession, exists := h.sessionMgr.GetSessionByUsername(targetUsername)
	if !exists {
		return sess.Send(protocol.NewErrorMessage(fmt.Sprintf("User '%s' is not online.", targetUsername)).Format())
	}

	// Send to target
	targetSession.Send(protocol.NewPrivateMessage(sess.GetUsername(), targetUsername, fmt.Sprintf("[PM] %s", message)).Format())

	// Confirm to sender
	sess.Send(protocol.NewCommandMessage(fmt.Sprintf("[PM to %s]: %s", targetUsername, message)).Format())

	return nil
}

// handleQuit disconnects the user
func (h *Handler) handleQuit(sess *session.Session) error {
	sess.Send(protocol.NewSystemMessage("Goodbye!").Format())
	return fmt.Errorf("user quit")
}
