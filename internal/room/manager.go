package room

import (
	"fmt"
	"sync"

	"github.com/mullayam/go-tcp-chat/internal/protocol"
	"github.com/mullayam/go-tcp-chat/internal/session"
)

// Manager manages all chat rooms
type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewManager creates a new room manager
func NewManager() *Manager {
	m := &Manager{
		rooms: make(map[string]*Room),
	}

	// Create default public room
	m.rooms[protocol.DefaultRoom] = NewRoom(protocol.DefaultRoom, TypePublic)

	return m
}

// GetRoom retrieves a room by name
func (m *Manager) GetRoom(name string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, exists := m.rooms[name]
	return room, exists
}

// CreateRoom creates a new private room
func (m *Manager) CreateRoom(name string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[name]; exists {
		return m.rooms[name], nil
	}

	room := NewRoom(name, TypePrivate)
	m.rooms[name] = room
	return room, nil
}

// JoinRoom adds a user to a room
func (m *Manager) JoinRoom(roomName string, session *session.Session) error {
	room, exists := m.GetRoom(roomName)
	if !exists {
		return fmt.Errorf("room '%s' does not exist", roomName)
	}

	room.AddMember(session)
	session.SetCurrentRoom(roomName)
	return nil
}

// LeaveRoom removes a user from a room
func (m *Manager) LeaveRoom(session *session.Session) {
	currentRoom := session.GetCurrentRoom()
	if currentRoom == "" {
		return
	}

	room, exists := m.GetRoom(currentRoom)
	if !exists {
		return
	}

	room.RemoveMember(session.GetUsername())
	session.SetCurrentRoom("")

	// Clean up empty private rooms (but not the default room)
	if room.Type == TypePrivate && room.GetMemberCount() == 0 {
		m.mu.Lock()
		delete(m.rooms, currentRoom)
		m.mu.Unlock()
	}
}

// GetDefaultRoom returns the default public room
func (m *Manager) GetDefaultRoom() *Room {
	room, _ := m.GetRoom(protocol.DefaultRoom)
	return room
}

// GetAllRoomNames returns a list of all room names
func (m *Manager) GetAllRoomNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.rooms))
	for name := range m.rooms {
		names = append(names, name)
	}
	return names
}

// BroadcastToRoom sends a message to all members of a room
func (m *Manager) BroadcastToRoom(roomName string, message *protocol.Message, excludeUsername string) error {
	room, exists := m.GetRoom(roomName)
	if !exists {
		return fmt.Errorf("room '%s' does not exist", roomName)
	}

	room.Broadcast(message, excludeUsername)
	return nil
}

// GetRoomInfo returns information about a room
func (m *Manager) GetRoomInfo(roomName string) (string, int, bool) {
	room, exists := m.GetRoom(roomName)
	if !exists {
		return "", 0, false
	}

	roomType := "public"
	if room.Type == TypePrivate {
		roomType = "private"
	}

	return roomType, room.GetMemberCount(), true
}
