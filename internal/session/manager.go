package session

import (
	"fmt"
	"net"
	"regexp"
	"sync"
)

// Manager manages all active sessions
type Manager struct {
	sessionsByIP       map[string]*Session
	sessionsByUsername map[string]*Session
	mu                 sync.RWMutex
	usernamePattern    *regexp.Regexp
	minUsernameLen     int
	maxUsernameLen     int
}

// NewManager creates a new session manager
func NewManager(minLen, maxLen int) *Manager {
	return &Manager{
		sessionsByIP:       make(map[string]*Session),
		sessionsByUsername: make(map[string]*Session),
		usernamePattern:    regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
		minUsernameLen:     minLen,
		maxUsernameLen:     maxLen,
	}
}

// AddSession adds a new session, enforcing one-connection-per-IP
func (m *Manager) AddSession(conn net.Conn, ip string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if IP already has an active session
	if _, exists := m.sessionsByIP[ip]; exists {
		return nil, fmt.Errorf("IP address %s already has an active connection", ip)
	}

	session := NewSession(conn, ip)
	m.sessionsByIP[ip] = session
	return session, nil
}

// RemoveSession removes a session and cleans up all references
func (m *Manager) RemoveSession(session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from IP map
	delete(m.sessionsByIP, session.IP)

	// Remove from username map if username was set
	if session.GetUsername() != "" {
		delete(m.sessionsByUsername, session.GetUsername())
	}
}

// ValidateUsername validates a username
func (m *Manager) ValidateUsername(username string) error {
	if len(username) < m.minUsernameLen {
		return fmt.Errorf("username must be at least %d characters", m.minUsernameLen)
	}
	if len(username) > m.maxUsernameLen {
		return fmt.Errorf("username must be at most %d characters", m.maxUsernameLen)
	}
	if !m.usernamePattern.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}
	return nil
}

// RegisterUsername registers a username for a session
func (m *Manager) RegisterUsername(session *Session, username string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if username is already taken
	if _, exists := m.sessionsByUsername[username]; exists {
		return fmt.Errorf("username '%s' is already taken", username)
	}

	// Register the username
	m.sessionsByUsername[username] = session
	session.SetUsername(username)
	return nil
}

// GetSessionByUsername retrieves a session by username
func (m *Manager) GetSessionByUsername(username string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessionsByUsername[username]
	return session, exists
}

// GetSessionByIP retrieves a session by IP
func (m *Manager) GetSessionByIP(ip string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessionsByIP[ip]
	return session, exists
}

// GetAllSessions returns all active sessions
func (m *Manager) GetAllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessionsByIP))
	for _, session := range m.sessionsByIP {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetAuthenticatedSessions returns all authenticated sessions
func (m *Manager) GetAuthenticatedSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, session := range m.sessionsByIP {
		if session.GetState() == StateAuthenticated {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetOnlineUsernames returns a list of all online usernames
func (m *Manager) GetOnlineUsernames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	usernames := make([]string, 0, len(m.sessionsByUsername))
	for username := range m.sessionsByUsername {
		usernames = append(usernames, username)
	}
	return usernames
}

// Count returns the number of active sessions
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessionsByIP)
}
