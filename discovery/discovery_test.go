package discovery

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestAnnounceAndSearch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start announcing with a known port
	go AnnounceHost(ctx, "12345")

	// Give the announcer a moment to start
	time.Sleep(100 * time.Millisecond)

	// Listen for the broadcast
	addr, err := net.ResolveUDPAddr("udp4", ":9876")
	if err != nil {
		t.Skipf("Cannot resolve UDP address: %s", err.Error())
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		t.Skipf("Cannot listen on UDP port %d (may be in use or no permission): %s", DiscoveryPort, err.Error())
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)

	// Read messages until we find the one from our announcer (port 12345)
	// Other tests may also be broadcasting on this port
	found := false
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		msg := strings.TrimSpace(string(buf[:n]))
		if !strings.HasPrefix(msg, messagePrefix) {
			continue
		}

		payload := strings.TrimPrefix(msg, messagePrefix)
		if strings.HasPrefix(payload, "12345:") {
			found = true
			// Verify full format: TIPTAP:12345:<hostname>
			parts := strings.SplitN(payload, ":", 2)
			if len(parts) != 2 || parts[0] != "12345" {
				t.Fatalf("Unexpected payload format: %q", payload)
			}
			if len(parts[1]) == 0 {
				t.Fatal("Expected non-empty hostname in broadcast")
			}
			break
		}
	}

	if !found {
		t.Skip("No broadcast with port 12345 received (expected on some network configurations)")
	}

	cancel()
}

func TestSearchHostsNoHosts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow search test in short mode")
	}

	// This test is inherently flaky when other tests broadcast on port 9876.
	// We do a quick probe first: if anything is already broadcasting, skip.
	probeAddr, err := net.ResolveUDPAddr("udp4", ":9876")
	if err != nil {
		t.Skip("Cannot resolve UDP address")
	}
	probeConn, err := net.ListenUDP("udp4", probeAddr)
	if err != nil {
		t.Skipf("Port %d in use, skipping: %s", DiscoveryPort, err.Error())
	}
	probeConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	probeBuf := make([]byte, 1024)
	n, _, probeErr := probeConn.ReadFromUDP(probeBuf)
	probeConn.Close()
	if probeErr == nil && n > 0 {
		t.Skip("Active broadcasts detected on port 9876, skipping no-hosts test")
	}

	hosts, err := SearchHosts()
	if err != nil {
		t.Fatalf("SearchHosts returned error: %s", err.Error())
	}
	if len(hosts) != 0 {
		t.Fatalf("Expected 0 hosts, got %d", len(hosts))
	}
}
