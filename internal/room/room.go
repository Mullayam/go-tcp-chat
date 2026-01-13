package room

import (
	"sync"

	"github.com/mullayam/go-tcp-chat/internal/protocol"
	"github.com/mullayam/go-tcp-chat/internal/session"
)

// Type represents the type of room
type Type int

const (
	// TypePublic represents a public room
	TypePublic Type = iota
	// TypePrivate represents a private room
	TypePrivate
)

// Room represents a chat room
type Room struct {
	Name    string
	Type    Type
	members map[string]*session.Session
	mu      sync.RWMutex
}

// NewRoom creates a new room
func NewRoom(name string, roomType Type) *Room {
	return &Room{
		Name:    name,
		Type:    roomType,
		members: make(map[string]*session.Session),
	}
}

// AddMember adds a member to the room
func (r *Room) AddMember(session *session.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[session.GetUsername()] = session
}

// RemoveMember removes a member from the room
func (r *Room) RemoveMember(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.members, username)
}

// HasMember checks if a user is a member
func (r *Room) HasMember(username string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.members[username]
	return exists
}

// GetMembers returns all members
func (r *Room) GetMembers() []*session.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members := make([]*session.Session, 0, len(r.members))
	for _, member := range r.members {
		members = append(members, member)
	}
	return members
}

// GetMemberCount returns the number of members
func (r *Room) GetMemberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.members)
}

// Broadcast sends a message to all members in the room
func (r *Room) Broadcast(message *protocol.Message, excludeUsername string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	formattedMsg := message.Format()
	for username, member := range r.members {
		if username != excludeUsername {
			_ = member.Send(formattedMsg)
		}
	}
}

// BroadcastToAll sends a message to all members including the sender
func (r *Room) BroadcastToAll(message *protocol.Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	formattedMsg := message.Format()
	for _, member := range r.members {
		_ = member.Send(formattedMsg)
	}
}

// GetMemberNames returns a list of member usernames
func (r *Room) GetMemberNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.members))
	for username := range r.members {
		names = append(names, username)
	}
	return names
}
