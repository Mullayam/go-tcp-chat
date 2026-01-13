package message

import (
	"strings"

	"github.com/mullayam/go-tcp-chat/internal/protocol"
	"github.com/mullayam/go-tcp-chat/internal/room"
	"github.com/mullayam/go-tcp-chat/internal/session"
)

// Router handles message routing
type Router struct {
	roomMgr *room.Manager
	handler *Handler
}

// NewRouter creates a new message router
func NewRouter(roomMgr *room.Manager, handler *Handler) *Router {
	return &Router{
		roomMgr: roomMgr,
		handler: handler,
	}
}

// Route routes a message from a user
func (r *Router) Route(sess *session.Session, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil
	}

	// Check if it's a command
	if strings.HasPrefix(message, "/") {
		return r.handler.HandleCommand(sess, message)
	}

	// It's a chat message - route based on context
	return r.routeChatMessage(sess, message)
}

// routeChatMessage routes a chat message based on the user's context
func (r *Router) routeChatMessage(sess *session.Session, content string) error {
	// Validate message length
	if len(content) > protocol.MaxMessageLength {
		return sess.Send(protocol.NewErrorMessage("Message too long. Maximum length is 1024 characters.").Format())
	}

	// Check if user is in a private chat
	privateChatWith := sess.GetPrivateChat()
	if privateChatWith != "" {
		// This would be for persistent private chat mode
		// For now, we only support /msg command for private messages
		sess.SetPrivateChat("")
	}

	// Route to current room
	currentRoom := sess.GetCurrentRoom()
	if currentRoom == "" {
		return sess.Send(protocol.NewErrorMessage("You are not in any room.").Format())
	}

	room, exists := r.roomMgr.GetRoom(currentRoom)
	if !exists {
		return sess.Send(protocol.NewErrorMessage("Current room no longer exists.").Format())
	}

	// Create and broadcast the message
	msg := protocol.NewChatMessage(sess.GetUsername(), content)
	room.BroadcastToAll(msg)

	return nil
}
