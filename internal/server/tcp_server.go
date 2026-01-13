package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/mullayam/go-tcp-chat/internal/auth"
	"github.com/mullayam/go-tcp-chat/internal/message"
	"github.com/mullayam/go-tcp-chat/internal/protocol"
	"github.com/mullayam/go-tcp-chat/internal/room"
	"github.com/mullayam/go-tcp-chat/internal/session"
)

// TCPServer represents the TCP chat server
type TCPServer struct {
	port         string
	sessionMgr   *session.Manager
	roomMgr      *room.Manager
	otpService   *auth.OTPService
	emailService *auth.EmailService
	router       *message.Router
	handler      *message.Handler
	listener     net.Listener
}

// NewTCPServer creates a new TCP server
func NewTCPServer(
	port string,
	sessionMgr *session.Manager,
	roomMgr *room.Manager,
	otpService *auth.OTPService,
	emailService *auth.EmailService,
) *TCPServer {
	handler := message.NewHandler(sessionMgr, roomMgr)
	router := message.NewRouter(roomMgr, handler)

	return &TCPServer{
		port:         port,
		sessionMgr:   sessionMgr,
		roomMgr:      roomMgr,
		otpService:   otpService,
		emailService: emailService,
		router:       router,
		handler:      handler,
	}
}

// Start starts the TCP server
func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener

	log.Printf("TCP Chat Server started on port %s", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// Stop stops the TCP server
func (s *TCPServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection handles a new client connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Extract IP address (without port)
	ip := s.extractIP(conn.RemoteAddr().String())
	log.Printf("New connection from %s", ip)

	// Try to add session (enforces one-connection-per-IP)
	sess, err := s.sessionMgr.AddSession(conn, ip)
	if err != nil {
		conn.Write([]byte(protocol.NewErrorMessage(err.Error()).Format()))
		log.Printf("Rejected connection from %s: %v", ip, err)
		return
	}

	// Ensure cleanup on disconnect
	defer s.cleanup(sess)

	// Send welcome message
	sess.Send(protocol.NewSystemMessage("Welcome to TCP Chat Server!").Format())
	sess.Send(protocol.NewSystemMessage("Please enter your email and authenticate to continue").Format())

	// Start authentication flow
	if err := s.authenticate(sess); err != nil {
		sess.Send(protocol.NewErrorMessage(fmt.Sprintf("Authentication failed: %v", err)).Format())
		log.Printf("Authentication failed for %s: %v", ip, err)
		return
	}

	// Join default room
	defaultRoom := s.roomMgr.GetDefaultRoom()
	defaultRoom.AddMember(sess)
	sess.SetCurrentRoom(protocol.DefaultRoom)

	// Notify user
	sess.Send(protocol.NewSystemMessage(fmt.Sprintf("You joined %s", protocol.DefaultRoom)).Format())
	sess.Send(protocol.NewSystemMessage("Type /help for available commands.").Format())

	// Notify room
	defaultRoom.Broadcast(protocol.NewSystemMessage(fmt.Sprintf("%s joined the room", sess.GetUsername())), sess.GetUsername())

	log.Printf("User %s authenticated from %s", sess.GetUsername(), ip)

	// Handle messages
	s.handleMessages(sess)
}

// authenticate handles the authentication flow
func (s *TCPServer) authenticate(sess *session.Session) error {
	// Request email
	sess.Send("\nEnter your email address: ")
	email, err := s.readLine(sess)
	if err != nil {
		return err
	}

	email = strings.TrimSpace(email)
	if !s.isValidEmail(email) {
		return fmt.Errorf("invalid email address")
	}

	sess.SetEmail(email)

	// Generate and send OTP
	otp, err := s.otpService.Generate(email)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	err = s.emailService.SendOTP(email, otp)
	if err != nil {
		s.otpService.Clear(email)
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	sess.SetState(session.StateAwaitingOTP)
	sess.Send(protocol.NewSystemMessage("OTP sent to your email. Please check your inbox.").Format())

	// Request OTP
	sess.Send("\nEnter OTP code: ")
	otpCode, err := s.readLine(sess)
	if err != nil {
		return err
	}

	otpCode = strings.TrimSpace(otpCode)

	// Validate OTP
	err = s.otpService.Validate(email, otpCode)
	if err != nil {
		return err
	}

	// Request username
	sess.Send("\nEnter username (3-16 characters, alphanumeric + underscore): ")
	username, err := s.readLine(sess)
	if err != nil {
		return err
	}

	username = strings.TrimSpace(username)

	// Validate username
	if err := s.sessionMgr.ValidateUsername(username); err != nil {
		return err
	}

	// Register username
	if err := s.sessionMgr.RegisterUsername(sess, username); err != nil {
		return err
	}

	sess.SetState(session.StateAuthenticated)
	return nil
}

// handleMessages handles incoming messages from a client
func (s *TCPServer) handleMessages(sess *session.Session) {
	for {
		line, err := s.readLine(sess)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from %s: %v", sess.GetUsername(), err)
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Route the message
		if err := s.router.Route(sess, line); err != nil {
			if err.Error() == "user quit" {
				return
			}
			log.Printf("Error routing message from %s: %v", sess.GetUsername(), err)
		}
	}
}

// readLine reads a line from the session
func (s *TCPServer) readLine(sess *session.Session) (string, error) {
	line, err := sess.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// cleanup cleans up a session on disconnect
func (s *TCPServer) cleanup(sess *session.Session) {
	username := sess.GetUsername()
	ip := sess.IP

	// Leave current room
	currentRoom := sess.GetCurrentRoom()
	if currentRoom != "" {
		s.roomMgr.LeaveRoom(sess)
		if room, exists := s.roomMgr.GetRoom(currentRoom); exists {
			if username != "" {
				room.Broadcast(protocol.NewSystemMessage(fmt.Sprintf("%s left the room", username)), "")
			}
		}
	}

	// Remove session
	s.sessionMgr.RemoveSession(sess)

	if username != "" {
		log.Printf("User %s disconnected from %s", username, ip)
	} else {
		log.Printf("Connection closed from %s", ip)
	}
}

// extractIP extracts the IP address from a remote address string
func (s *TCPServer) extractIP(remoteAddr string) string {
	// Remove port
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}

// isValidEmail validates an email address
func (s *TCPServer) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
