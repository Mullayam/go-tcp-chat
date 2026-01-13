package room

import (
	"sync"
	"time"

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

// HistoryItem represents a stored message
type HistoryItem struct {
	Content   string
	Timestamp time.Time
}

// Room represents a chat room
type Room struct {
	Name    string
	Type    Type
	members map[string]*session.Session
	history []HistoryItem // Store recent messages
	mu      sync.RWMutex
}

// NewRoom creates a new room
func NewRoom(name string, roomType Type) *Room {
	return &Room{
		Name:    name,
		Type:    roomType,
		members: make(map[string]*session.Session),
		history: make([]HistoryItem, 0),
	}
}

// AddMember adds a member to the room and sends history
func (r *Room) AddMember(session *session.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanupHistory() // Lazy cleanup before adding member

	// Replay history to the new member
	if len(r.history) > 0 {
		_ = session.Send(protocol.NewSystemMessage("--- History (last 5 min) ---").Format())
		for _, item := range r.history {
			_ = session.Send(item.Content)
		}
		_ = session.Send(protocol.NewSystemMessage("----------------------------").Format())
	}

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
	r.mu.Lock() // Upgraded to Lock for history modification
	defer r.mu.Unlock()

	formattedMsg := message.Format()

	// Store in history
	r.addToHistory(formattedMsg)

	for username, member := range r.members {
		if username != excludeUsername {
			_ = member.Send(formattedMsg)
		}
	}
}

// BroadcastToAll sends a message to all members including the sender
func (r *Room) BroadcastToAll(message *protocol.Message) {
	r.mu.Lock() // Upgraded to Lock for history modification
	defer r.mu.Unlock()

	formattedMsg := message.Format()

	// Store in history
	r.addToHistory(formattedMsg)

	for _, member := range r.members {
		_ = member.Send(formattedMsg)
	}
}

// addToHistory adds a message to history and performs cleanup
func (r *Room) addToHistory(content string) {
	r.history = append(r.history, HistoryItem{
		Content:   content,
		Timestamp: time.Now(),
	})
	r.cleanupHistory()
}

// cleanupHistory removes messages older than 5 minutes
// Caller must hold the lock
func (r *Room) cleanupHistory() {
	cutoff := time.Now().Add(-5 * time.Minute)

	if len(r.history) > 0 {
		if r.history[len(r.history)-1].Timestamp.Before(cutoff) {
			// All items are old so clear everything
			r.history = make([]HistoryItem, 0)
			return
		}

		if r.history[0].Timestamp.Before(cutoff) {
			// Prune old messages
			newHistory := make([]HistoryItem, 0, len(r.history))
			for _, item := range r.history {
				if item.Timestamp.After(cutoff) {
					newHistory = append(newHistory, item)
				}
			}
			r.history = newHistory
		}
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
