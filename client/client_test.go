package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestMakeEmptyClient verifies that makeEmptyClient returns a zero-value struct
// with a nil connection and a default Logger.
func TestMakeEmptyClient(t *testing.T) {
	c := makeEmptyClient()

	if c.connection != nil {
		t.Errorf("expected connection to be nil, got %v", c.connection)
	}

	// Logger is a value type; the zero value has both internal fields nil.
	// Calling Logerr on a zero Logger should not panic (it lazily initialises).
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("calling Logerr on zero Logger panicked: %v", r)
		}
	}()
	c.logger.Logerr("test")
}

// TestStartClientConnectionSuccess verifies that the returned client's
// remote address matches the listener's address.
func TestStartClientConnectionSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := fmt.Sprintf("%d", addr.Port)

	client, err := startClientConnection("127.0.0.1", port)
	if err != nil {
		t.Fatalf("startClientConnection returned error: %v", err)
	}
	defer client.connection.Close()

	if client.connection == nil {
		t.Fatal("expected non-nil connection after successful connect")
	}

	remoteAddr := client.connection.RemoteAddr().String()
	if !strings.HasSuffix(remoteAddr, ":"+port) {
		t.Errorf("expected remote address to end with :%s, got %s", port, remoteAddr)
	}
}

// TestConnectClientFullExchange starts a mock TCP server that performs the
// welcome/prompt/username handshake, then sends one chat message. Stdin is
// redirected via os.Pipe so ConnectClient can read the username without
// blocking on the real terminal. The test verifies the exchange completes
// without error.
func TestConnectClientFullExchange(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := fmt.Sprintf("%d", addr.Port)

	// serverDone carries any server-side error back to the test goroutine.
	serverDone := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- fmt.Errorf("accept: %w", err)
			return
		}
		defer conn.Close()

		// Handshake: send welcome line then prompt line.
		conn.Write([]byte("Welcome to Tip-Tap!\n"))
		conn.Write([]byte("Enter your username:\n"))

		// Read the username the client sends back.
		reader := bufio.NewReader(conn)
		_, err = reader.ReadString('\n')
		if err != nil {
			serverDone <- fmt.Errorf("reading username: %w", err)
			return
		}

		// Send one chat message so receiveLoop has something to process.
		conn.Write([]byte("alice: hello\n"))

		// Close the connection so both the receive goroutine and the stdin
		// loop in ConnectClient terminate cleanly.
		serverDone <- nil
	}()

	// Redirect stdin so ConnectClient reads "testuser\n" instead of blocking
	// on the real terminal.
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	stdinW.WriteString("testuser\n")
	stdinW.Close()

	origStdin := os.Stdin
	os.Stdin = stdinR
	defer func() {
		os.Stdin = origStdin
		stdinR.Close()
	}()

	// ConnectClient blocks until the connection closes; run it in a goroutine
	// with a timeout guard.
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- ConnectClient("127.0.0.1", port)
	}()

	// Wait for the mock server to finish its side of the exchange first.
	select {
	case err := <-serverDone:
		if err != nil {
			t.Errorf("server goroutine error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server goroutine")
	}

	// Now wait for ConnectClient to return.
	select {
	case err := <-clientDone:
		if err != nil {
			t.Errorf("ConnectClient returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for ConnectClient to return")
	}
}

// TestReceiveLoopFormatsMessages verifies that receiveLoop wraps the sender
// name in parentheses when a message contains ": ".
func TestReceiveLoopFormatsMessages(t *testing.T) {
	// net.Pipe gives two synchronous, in-memory net.Conn endpoints.
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Capture stdout so we can inspect what receiveLoop prints.
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe for stdout: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = stdoutW
	defer func() { os.Stdout = origStdout }()

	done := make(chan struct{})
	var termMu sync.Mutex
	reader := bufio.NewReader(clientConn)
	go receiveLoop(clientConn, reader, done, "[me]: ", &termMu)

	// Write a formatted message followed by one that has no ": " separator.
	serverConn.Write([]byte("alice: hello world\n"))
	serverConn.Write([]byte("no-separator-line\n"))
	// Close server side to make receiveLoop exit.
	serverConn.Close()

	// Wait for receiveLoop to finish.
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for receiveLoop to exit")
	}

	// Restore stdout and read captured output.
	stdoutW.Close()
	os.Stdout = origStdout

	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := stdoutR.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	stdoutR.Close()
	output := sb.String()

	if !strings.Contains(output, "(alice): hello world") {
		t.Errorf("expected output to contain \"(alice): hello world\", got: %q", output)
	}

	if !strings.Contains(output, "no-separator-line") {
		t.Errorf("expected output to contain \"no-separator-line\" verbatim, got: %q", output)
	}

	if strings.Contains(output, "(no-separator-line)") {
		t.Errorf("lines without \": \" should not be wrapped in parentheses, got: %q", output)
	}
}

func TestStartClientConnection(t *testing.T) {
	// Start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %s", err.Error())
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.AddrPort().Port()

	client, err := startClientConnection("127.0.0.1", fmt.Sprintf("%d", port))
	if err != nil {
		t.Fatalf("Failed to connect client: %s", err.Error())
	}
	defer client.connection.Close()

	if client.connection == nil {
		t.Fatal("Client connection is nil after successful connect")
	}
}

func TestStartClientConnectionFailure(t *testing.T) {
	// Try to connect to a port that is not listening
	_, err := startClientConnection("127.0.0.1", "1")
	if err == nil {
		t.Fatal("Expected error when connecting to non-listening port, got nil")
	}
}
