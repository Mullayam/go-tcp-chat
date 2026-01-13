package session

import (
	"bufio"
	"net"
	"sync"
)

// State represents the authentication state of a session
type State int

const (
	// StateUnauthenticated means the user hasn't authenticated yet
	StateUnauthenticated State = iota
	// StateAwaitingOTP means OTP has been sent, waiting for verification
	StateAwaitingOTP
	// StateAuthenticated means the user is fully authenticated
	StateAuthenticated
)

// Session represents a user session
type Session struct {
	Username string
	Email    string
	IP       string
	State    State
	Conn     net.Conn
	Writer   *bufio.Writer
	Reader   *bufio.Reader

	// Current context
	CurrentRoom     string
	PrivateChatWith string

	mu sync.RWMutex
}

// NewSession creates a new session
func NewSession(conn net.Conn, ip string) *Session {
	return &Session{
		IP:     ip,
		State:  StateUnauthenticated,
		Conn:   conn,
		Writer: bufio.NewWriter(conn),
		Reader: bufio.NewReader(conn),
	}
}

// Send sends a message to the client
func (s *Session) Send(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.Writer.WriteString(message)
	if err != nil {
		return err
	}
	return s.Writer.Flush()
}

// SetState sets the session state
func (s *Session) SetState(state State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State = state
}

// GetState gets the session state
func (s *Session) GetState() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

// SetUsername sets the username
func (s *Session) SetUsername(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Username = username
}

// GetUsername gets the username
func (s *Session) GetUsername() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Username
}

// SetEmail sets the email
func (s *Session) SetEmail(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Email = email
}

// SetCurrentRoom sets the current room
func (s *Session) SetCurrentRoom(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentRoom = room
}

// GetCurrentRoom gets the current room
func (s *Session) GetCurrentRoom() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentRoom
}

// SetPrivateChat sets the private chat partner
func (s *Session) SetPrivateChat(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PrivateChatWith = username
}

// GetPrivateChat gets the private chat partner
func (s *Session) GetPrivateChat() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PrivateChatWith
}

// Close closes the session connection
func (s *Session) Close() error {
	return s.Conn.Close()
}
