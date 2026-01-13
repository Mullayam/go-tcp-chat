package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	// Connect to server
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to TCP Chat Server")
	fmt.Println("=====================================")

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
					fmt.Printf("\nConnection closed: %v\n", err)
				} else {
					fmt.Println("\nServer disconnected.")
				}
				os.Exit(0)
			}
			// Print without adding extra newline (server messages already have \n)
			fmt.Print(line)
		}
	}()

	// Read from stdin and send to server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		// Send to server
		_, err := fmt.Fprintf(conn, "%s\n", line)
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
			break
		}

		// Check if user wants to quit
		if strings.TrimSpace(line) == "/quit" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}

	wg.Wait()
}
