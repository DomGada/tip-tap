package discovery

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	DiscoveryPort     = 9876
	SearchTimeout     = 5 * time.Second
	BroadcastInterval = 2 * time.Second
	broadcastAddr     = "255.255.255.255"
	messagePrefix     = "TIPTAP:"
)

// AnnounceHost broadcasts the host's presence on the LAN via UDP until ctx is cancelled.
func AnnounceHost(ctx context.Context, tcpPort string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	msg := fmt.Sprintf("%s%s:%s\n", messagePrefix, tcpPort, hostname)

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcastAddr, DiscoveryPort))
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn.Write([]byte(msg))
			select {
			case <-ctx.Done():
				return
			case <-time.After(BroadcastInterval):
			}
		}
	}
}

// SearchHosts listens for host announcements on the LAN and returns discovered hosts.
func SearchHosts() ([]HostInfo, error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP port %d: %w", DiscoveryPort, err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(SearchTimeout))

	seen := make(map[string]HostInfo)
	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			// Timeout or other error — return what we have
			break
		}

		msg := strings.TrimSpace(string(buf[:n]))
		if !strings.HasPrefix(msg, messagePrefix) {
			continue
		}

		payload := strings.TrimPrefix(msg, messagePrefix)
		parts := strings.SplitN(payload, ":", 2)
		if len(parts) != 2 {
			continue
		}

		tcpPort := parts[0]
		hostName := parts[1]
		ip := remoteAddr.IP.String()
		key := ip + ":" + tcpPort

		if _, exists := seen[key]; !exists {
			seen[key] = HostInfo{
				Name:    hostName,
				IP:      ip,
				TCPPort: tcpPort,
			}
		}
	}

	hosts := make([]HostInfo, 0, len(seen))
	for _, h := range seen {
		hosts = append(hosts, h)
	}
	return hosts, nil
}
