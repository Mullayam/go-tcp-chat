package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"
	ColorOrange  = "\033[38;5;208m"
)

func main() {
	// Connect to server
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Printf("%sFailed to connect: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println(ColorCyan + "Connected to TCP Chat Server" + ColorReset)
	fmt.Println(ColorCyan + "=====================================" + ColorReset)

	var wg sync.WaitGroup
	wg.Add(1)

	// Read from server in a goroutine
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("\n%sConnection closed: %v%s\n", ColorRed, err, ColorReset)
				} else {
					fmt.Printf("\n%sServer disconnected.%s\n", ColorYellow, ColorReset)
				}
				os.Exit(0)
			}

			// Process and display the line
			printServerMessage(line)
		}
	}()

	// Wait a bit for initial welcome messages to ensure they print before the first prompt
	time.Sleep(200 * time.Millisecond)

	// Read from stdin and send to server
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Show input prompt "C -> "
		fmt.Printf("%sC -> %s", ColorGreen+ColorBold, ColorReset)

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		// Move cursor up one line and clear it to remove local echo
		// (The server will broadcast messages back, preventing double-lines)
		fmt.Print("\033[1A\033[2K")

		// Send to server
		_, err := fmt.Fprintf(conn, "%s\n", line)
		if err != nil {
			fmt.Printf("%sFailed to send message: %v%s\n", ColorRed, err, ColorReset)
			break
		}

		// Check if user wants to quit
		if strings.TrimSpace(line) == "/quit" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("%sError reading input: %v%s\n", ColorRed, err, ColorReset)
	}

	wg.Wait()
}

func printServerMessage(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	// Chat messages
	// [username]: message
	if strings.Contains(line, "]:") {
		parts := strings.SplitN(line, "]:", 2)
		if len(parts) == 2 {
			usernamePart := parts[0] + "]:" // [user]:
			messagePart := parts[1]

			// Extract username for color hashing
			cleanName := strings.TrimPrefix(usernamePart, "[")
			cleanName = strings.TrimSuffix(cleanName, "]:")

			// Highlight PMs specially
			if strings.Contains(usernamePart, "PM") {
				// Orange color for PMs
				fmt.Printf("\r\033[K%s%s%s%s\n", ColorOrange+ColorBold, usernamePart, ColorReset, messagePart)
				return
			}

			// Regular chat message
			// User: Green/Blue/etc (hashed), Message: White/Bright
			userColor := getUsernameColor(cleanName)
			fmt.Printf("\r\033[K%s%s%s%s\n", userColor+ColorBold, usernamePart, ColorReset, messagePart)
			return
		}
	}

	// Clean up system messages (Remove S -> prefix requirement)
	// Server sends "*** Content ***" via NewSystemMessage
	if strings.HasPrefix(line, "***") && strings.HasSuffix(line, "***") {
		content := strings.Trim(line, "* ")
		// System messages in Yellow Bold (No S -> prefix)
		fmt.Printf("\r\033[K%s%s%s\n", ColorYellow+ColorBold, content, ColorReset)
		return
	}

	// Error messages
	if strings.HasPrefix(line, "ERROR:") {
		content := strings.TrimPrefix(line, "ERROR: ")
		fmt.Printf("\r\033[K%sERROR: %s%s\n", ColorRed+ColorBold, content, ColorReset)
		return
	}

	// Default fallback (Command responses, etc)
	if !strings.HasPrefix(line, "[") {
		// Just print the line without S -> prefix, using Yellow for server text
		fmt.Printf("\r\033[K%s%s%s\n", ColorYellow, line, ColorReset)
	} else {
		fmt.Printf("\r\033[K%s\n", line)
	}
}

func getUsernameColor(username string) string {
	hash := 0
	for _, ch := range username {
		hash += int(ch)
	}
	// Palette: Green, Blue, Magenta, Cyan, Bright Blue
	colors := []string{ColorGreen, ColorBlue, ColorMagenta, ColorCyan, "\033[94m"}
	return colors[hash%len(colors)]
}
