package protocol

import "fmt"

// MessageType represents the type of message
type MessageType int

const (
	// MessageTypeSystem represents system messages (join, leave, errors)
	MessageTypeSystem MessageType = iota
	// MessageTypeChat represents user chat messages
	MessageTypeChat
	// MessageTypeError represents error messages
	MessageTypeError
	// MessageTypeCommand represents command responses
	MessageTypeCommand
)

// Message represents a formatted message
type Message struct {
	Type    MessageType
	From    string
	Content string
	To      string // For private messages
}

// Format formats a message for display to the client
func (m *Message) Format() string {
	switch m.Type {
	case MessageTypeSystem:
		return fmt.Sprintf("*** %s ***\n", m.Content)
	case MessageTypeChat:
		if m.From != "" {
			return fmt.Sprintf("[%s]: %s\n", m.From, m.Content)
		}
		return fmt.Sprintf("%s\n", m.Content)
	case MessageTypeError:
		return fmt.Sprintf("ERROR: %s\n", m.Content)
	case MessageTypeCommand:
		return fmt.Sprintf("%s\n", m.Content)
	default:
		return fmt.Sprintf("%s\n", m.Content)
	}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(content string) *Message {
	return &Message{
		Type:    MessageTypeSystem,
		Content: content,
	}
}

// NewChatMessage creates a new chat message
func NewChatMessage(from, content string) *Message {
	return &Message{
		Type:    MessageTypeChat,
		From:    from,
		Content: content,
	}
}

// NewPrivateMessage creates a new private message
func NewPrivateMessage(from, to, content string) *Message {
	return &Message{
		Type:    MessageTypeChat,
		From:    from,
		To:      to,
		Content: content,
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(content string) *Message {
	return &Message{
		Type:    MessageTypeError,
		Content: content,
	}
}

// NewCommandMessage creates a new command response message
func NewCommandMessage(content string) *Message {
	return &Message{
		Type:    MessageTypeCommand,
		Content: content,
	}
}

// Protocol constants
const (
	// DefaultRoom is the default public room
	DefaultRoom = "#general"

	// MaxMessageLength is the maximum length of a message
	MaxMessageLength = 1024

	// MaxUsernameLength is the maximum length of a username
	MaxUsernameLength = 16

	// MinUsernameLength is the minimum length of a username
	MinUsernameLength = 3
)
