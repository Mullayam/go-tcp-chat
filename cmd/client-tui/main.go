package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	senderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))   // Magenta
	systemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))   // Yellow
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // Red
	pmStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange

	usernameStyles = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // Green
		lipgloss.NewStyle().Foreground(lipgloss.Color("4")), // Blue
		lipgloss.NewStyle().Foreground(lipgloss.Color("6")), // Cyan
		lipgloss.NewStyle().Foreground(lipgloss.Color("5")), // Magenta
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")), // Yellow
	}
)

type errMsg error

type serverMsg string

type model struct {
	viewport  viewport.Model
	textInput textinput.Model
	messages  []string
	conn      net.Conn
	msgChan   chan string
	err       error
	roomName  string
	width     int
	height    int
	ready     bool
}

func initialModel(conn net.Conn) model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 30

	// Channel for incoming messages
	msgChan := make(chan string)

	// Start reading goroutine
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				msgChan <- "ERROR: Connection lost"
				close(msgChan)
				return
			}
			msgChan <- strings.TrimRight(line, "\r\n")
		}
	}()

	return model{
		textInput: ti,
		messages:  []string{},
		conn:      conn,
		msgChan:   msgChan,
		roomName:  "#general",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		waitForServerMsg(m.msgChan),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		headerHeight := 2 // Border + Room Name
		footerHeight := 3 // Border + Input
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight + 1
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
		m.textInput.Width = msg.Width - 4

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() == "" {
				return m, nil
			}
			// Send message
			fmt.Fprintf(m.conn, "%s\n", m.textInput.Value())

			// Commands
			if strings.TrimSpace(m.textInput.Value()) == "/quit" {
				return m, tea.Quit
			}

			// We don't append local message, we rely on server echo for history
			m.textInput.SetValue("")
		}

	case serverMsg:
		content := string(msg)

		// Parse room name
		if strings.Contains(content, "You joined") {
			parts := strings.Split(content, "You joined ")
			if len(parts) > 1 {
				m.roomName = strings.Trim(parts[1], "* ")
			}
		}

		// Parse colors and styling
		styledContent := styleMessage(content)
		m.messages = append(m.messages, styledContent)

		// Truncate history if massive
		if len(m.messages) > 1000 {
			m.messages = m.messages[len(m.messages)-1000:]
		}

		if m.ready {
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		}

		return m, waitForServerMsg(m.msgChan)

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	borderColor := lipgloss.Color("36") // Cyan

	// Header
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Pinkish
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Align(lipgloss.Center).
		Render("Room: " + m.roomName)

	// Footer (Input)
	footer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Render(m.textInput.View())

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

func waitForServerMsg(sub chan string) tea.Cmd {
	return func() tea.Msg {
		val, ok := <-sub
		if !ok {
			return tea.Quit
		}
		return serverMsg(val)
	}
}

func styleMessage(line string) string {
	// Re-implement the coloring logic for Bubble Tea

	// System Messages
	if strings.Contains(line, "Welcome") || strings.Contains(line, "joined") ||
		strings.Contains(line, "History") || strings.HasPrefix(line, "***") {
		clean := strings.Trim(line, "*")
		return systemStyle.Render(clean)
	}

	// Error
	if strings.HasPrefix(strings.ToUpper(line), "ERROR:") {
		return errorStyle.Render(line)
	}

	// [user]: msg
	if strings.Contains(line, "]:") {
		parts := strings.SplitN(line, "]:", 2)
		if len(parts) == 2 {
			userPart := parts[0] + "]:"
			msgPart := parts[1]

			cleanName := strings.Trim(parts[0], "[]")

			// PM check
			if strings.Contains(parts[0], "PM") {
				return pmStyle.Render(userPart) + msgPart
			}

			colStyle := getUsernameColorStyle(cleanName)
			return colStyle.Render(userPart) + msgPart
		}
	}

	return line // Default
}

func getUsernameColorStyle(username string) lipgloss.Style {
	hash := 0
	for _, ch := range username {
		hash += int(ch)
	}
	return usernameStyles[hash%len(usernameStyles)]
}

func main() {
	// Parse flags
	urlFlag := flag.String("url", "", "Connection URL (e.g., enjoys://tcp-chat@127.0.0.1:8888)")
	flag.Parse()

	address := "localhost:8888"
	if *urlFlag != "" {
		cleanUrl := strings.TrimSpace(*urlFlag)
		if strings.HasPrefix(cleanUrl, "enjoys://tcp-chat@") {
			address = strings.TrimPrefix(cleanUrl, "enjoys://tcp-chat@")
		} else {
			fmt.Println("Invalid URL format.")
			os.Exit(1)
		}
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Could not connect:", err)
		os.Exit(1)
	}
	defer conn.Close()

	if _, err := tea.NewProgram(initialModel(conn), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
