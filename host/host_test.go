package host

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestHostInputValid(t *testing.T) {

}

func TestCreateHost(t *testing.T) {
	host, err := StartHostConnection("25565")
	if err != nil {
		t.Fatalf("TestCreateHost failed: %s\n", err.Error())
	}
	StopHostConnection(host)
}

// TestStopHostConnection verifies that StopHostConnection closes the listener
// and returns true on success.
func TestStopHostConnection(t *testing.T) {
	host, err := StartHostConnection("25566")
	if err != nil {
		t.Fatalf("failed to start host: %v", err)
	}

	stopped := StopHostConnection(host)
	if !stopped {
		t.Fatal("StopHostConnection returned false, expected true")
	}

	// A second attempt on the already-closed listener must return false.
	stopped = StopHostConnection(host)
	if stopped {
		t.Fatal("StopHostConnection returned true on already-closed listener, expected false")
	}
}

// TestCombineIPAndPort verifies that combineIPAndPort formats the address correctly.
func TestCombineIPAndPort(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ip       string
		port     string
		expected string
	}{
		{"", "8080", ":8080"},
		{"127.0.0.1", "9000", "127.0.0.1:9000"},
		{"0.0.0.0", "25565", "0.0.0.0:25565"},
		{"localhost", "1234", "localhost:1234"},
	}

	for _, tc := range cases {
		result := combineIPAndPort(tc.ip, tc.port)
		if result != tc.expected {
			t.Errorf("combineIPAndPort(%q, %q) = %q, want %q", tc.ip, tc.port, result, tc.expected)
		}
	}
}

// freePort asks the OS for an available TCP port by listening on :0,
// capturing the address, then closing the listener before returning.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("could not obtain free port: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	l.Close()
	return fmt.Sprintf("%d", addr.AddrPort().Port())
}

// readLineTimeout reads one newline-terminated line from r with a deadline.
func readLineTimeout(t *testing.T, r *bufio.Reader, conn net.Conn, timeout time.Duration) string {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	line, err := r.ReadString('\n')
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		t.Fatalf("timed out or error reading line: %v", err)
	}
	return strings.TrimRight(line, "\n")
}

// readUntilContains reads lines until one contains the target string, skipping
// join/leave announcements and other irrelevant messages.
func readUntilContains(t *testing.T, r *bufio.Reader, conn net.Conn, target string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(deadline)
		line, err := r.ReadString('\n')
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			t.Fatalf("error reading while looking for %q: %v", target, err)
		}
		line = strings.TrimRight(line, "\n")
		if strings.Contains(line, target) {
			return line
		}
	}
	t.Fatalf("timed out looking for line containing %q", target)
	return ""
}

// drainWelcome reads and discards the two welcome lines the host sends on connect.
func drainWelcome(t *testing.T, r *bufio.Reader, conn net.Conn) {
	t.Helper()
	readLineTimeout(t, r, conn, 2*time.Second) // "Welcome to the TipTap Host!"
	readLineTimeout(t, r, conn, 2*time.Second) // "Please enter your display name :3"
}

// TestHostClientIntegration starts a host with a context, connects two clients,
// has one send a message, and verifies the other receives the broadcast.
func TestHostClientIntegration(t *testing.T) {
	port := freePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run the host in a background goroutine; it exits when ctx is cancelled.
	hostErr := make(chan error, 1)
	go func() {
		hostErr <- RunHost(ctx, port)
	}()

	// Give the host a moment to start listening.
	time.Sleep(50 * time.Millisecond)

	// Connect client A.
	connA, err := net.DialTimeout("tcp", ":"+port, 2*time.Second)
	if err != nil {
		t.Fatalf("client A dial failed: %v", err)
	}
	defer connA.Close()
	readerA := bufio.NewReader(connA)
	drainWelcome(t, readerA, connA)

	// Send username for client A.
	connA.Write([]byte("Alice\n"))

	// Connect client B.
	connB, err := net.DialTimeout("tcp", ":"+port, 2*time.Second)
	if err != nil {
		t.Fatalf("client B dial failed: %v", err)
	}
	defer connB.Close()
	readerB := bufio.NewReader(connB)
	drainWelcome(t, readerB, connB)

	// Send username for client B.
	connB.Write([]byte("Bob\n"))

	// Allow connectionLoop goroutines to process the usernames.
	time.Sleep(50 * time.Millisecond)

	// Client A sends a message.
	connA.Write([]byte("hello from Alice\n"))

	// Client B should receive "Alice: hello from Alice" (may be preceded by join announcements).
	line := readUntilContains(t, readerB, connB, "hello from Alice", 3*time.Second)
	if !strings.Contains(line, "Alice") {
		t.Errorf("client B received unexpected message: %q", line)
	}

	// Cancel the host and verify it shuts down cleanly.
	cancel()
	select {
	case err := <-hostErr:
		if err != nil {
			t.Errorf("RunHost returned non-nil error on clean shutdown: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("RunHost did not shut down within 3 seconds after context cancel")
	}
}

// TestBroadcastMessage verifies that broadcastMessage sends to all connections
// except the sender.
func TestBroadcastMessage(t *testing.T) {
	// Reset global connections before and after this test.
	connMutex.Lock()
	connections = nil
	connMutex.Unlock()
	defer func() {
		connMutex.Lock()
		connections = nil
		connMutex.Unlock()
	}()

	// Create three in-process pipe pairs acting as fake TCP connections.
	// senderServer is what the host holds; senderClient is the sender side.
	senderServer, _ := net.Pipe()
	recvServer1, recvClient1 := net.Pipe()
	recvServer2, recvClient2 := net.Pipe()

	defer senderServer.Close()
	defer recvServer1.Close()
	defer recvServer2.Close()
	defer recvClient1.Close()
	defer recvClient2.Close()

	// Register all three server-side ends in the global connections slice.
	addConnection(senderServer)
	addConnection(recvServer1)
	addConnection(recvServer2)

	// Collect what each receiver reads.
	got1 := make(chan string, 1)
	got2 := make(chan string, 1)

	readOne := func(client net.Conn, ch chan<- string) {
		client.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 256)
		n, _ := client.Read(buf)
		ch <- string(buf[:n])
	}
	go readOne(recvClient1, got1)
	go readOne(recvClient2, got2)

	broadcastMessage(senderServer, "test broadcast")

	for i, ch := range []chan string{got1, got2} {
		select {
		case msg := <-ch:
			if !strings.Contains(msg, "test broadcast") {
				t.Errorf("receiver %d got unexpected message: %q", i+1, msg)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("receiver %d did not receive broadcast within timeout", i+1)
		}
	}

	// The sender itself must not have received anything.
	senderServer.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buf := make([]byte, 256)
	n, _ := senderServer.Read(buf)
	if n > 0 {
		t.Errorf("sender unexpectedly received its own broadcast: %q", string(buf[:n]))
	}
}

// TestMultipleClients connects 3 clients through a real host, has one send a
// message, and verifies both other clients receive it.
func TestMultipleClients(t *testing.T) {
	port := freePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go RunHost(ctx, port)
	time.Sleep(50 * time.Millisecond)

	type client struct {
		conn   net.Conn
		reader *bufio.Reader
		name   string
	}

	dialClient := func(name string) client {
		conn, err := net.DialTimeout("tcp", ":"+port, 2*time.Second)
		if err != nil {
			t.Fatalf("dial failed for %s: %v", name, err)
		}
		r := bufio.NewReader(conn)
		drainWelcome(t, r, conn)
		conn.Write([]byte(name + "\n"))
		return client{conn, r, name}
	}

	alice := dialClient("Alice")
	bob := dialClient("Bob")
	carol := dialClient("Carol")
	defer alice.conn.Close()
	defer bob.conn.Close()
	defer carol.conn.Close()

	// Allow all connectionLoop goroutines to finish reading usernames.
	time.Sleep(100 * time.Millisecond)

	// Alice sends a message.
	alice.conn.Write([]byte("hi everyone\n"))

	const want = "hi everyone"
	for _, c := range []client{bob, carol} {
		line := readUntilContains(t, c.reader, c.conn, want, 3*time.Second)
		if !strings.Contains(line, want) {
			t.Errorf("client %s received %q, expected it to contain %q", c.name, line, want)
		}
	}
}
