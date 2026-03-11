package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	CONNTYPE = "tcp"
)

func ConnectClient(ip_addr, port string) error {
	client, err := startClientConnection(ip_addr, port)
	if err != nil {
		return fmt.Errorf("unable to form connection with host: %w", err)
	}

	conn := client.connection
	connReader := bufio.NewReader(conn)

	// Synchronous username exchange: read welcome + prompt, send username
	welcome, err := connReader.ReadString('\n')
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read welcome message: %w", err)
	}
	fmt.Print(welcome)

	prompt, err := connReader.ReadString('\n')
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read username prompt: %w", err)
	}
	fmt.Print(prompt)

	stdinReader := bufio.NewReader(os.Stdin)
	fmt.Print("Name: ")
	username, _ := stdinReader.ReadString('\n')
	displayName := username[:len(username)-1] // trim newline for prompt display
	_, err = conn.Write([]byte(username))
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to send username: %w", err)
	}

	// Spawn receive goroutine
	inputPrompt := fmt.Sprintf("[%s]: ", displayName)
	var termMu sync.Mutex
	done := make(chan struct{})
	go receiveLoop(conn, connReader, done, inputPrompt, &termMu)

	// Main goroutine: read stdin, write to server
	termMu.Lock()
	fmt.Print(inputPrompt)
	termMu.Unlock()
	for {
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			break
		}
		_, writeErr := conn.Write([]byte(input))
		if writeErr != nil {
			break
		}
		termMu.Lock()
		fmt.Print(inputPrompt)
		termMu.Unlock()
	}

	conn.Close()
	<-done
	return nil
}

func receiveLoop(conn net.Conn, reader *bufio.Reader, done chan struct{}, prompt string, termMu *sync.Mutex) {
	defer close(done)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		// Clear current line, print incoming message, then reprint the prompt
		formatted := msg
		if idx := strings.Index(msg, ": "); idx > 0 {
			name := msg[:idx]
			rest := msg[idx:]
			formatted = fmt.Sprintf("(%s)%s", name, rest)
		}
		termMu.Lock()
		// \r returns to line start, \033[K clears to end of line
		fmt.Printf("\r\033[K%s%s", formatted, prompt)
		termMu.Unlock()
	}
}

func startClientConnection(ip_addr, port string) (TipTapClient, error) {
	client := makeEmptyClient()
	logger := client.logger
	conn, err := net.Dial(CONNTYPE, ip_addr+":"+port)
	if err != nil {
		logger.Logerr("Failed to connect: " + err.Error())
		return client, err
	}

	client.connection = conn
	return client, nil

}
