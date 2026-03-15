package host

/*
This defines our public library functions.
See host_util.go for utility functions
*/

import (
	"context"
	"fmt"
	"net"

	"github.com/DomGada/Tip-Tap/discovery"
)

const (
	CONNHOST = ""
	CONNTYPE = "tcp"
) // TODO: we will want to make this more functional lol

func RunHost(ctx context.Context, port string) error {
	hostConn, err := StartHostConnection(port)
	if err != nil {
		return fmt.Errorf("unable to form connection with host: %w", err)
	}
	logger := hostConn.logger
	hostSock := hostConn.listener
	defer hostSock.Close()

	// Start UDP discovery broadcast
	go discovery.AnnounceHost(ctx, port)

	// Close listener when context is cancelled
	go func() {
		<-ctx.Done()
		hostSock.Close()
	}()

	for {
		fmt.Printf("Waiting for additional clients to connect...\n")
		c, err := hostSock.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				logger.Loginfo("Host shutting down.")
				closeAllConnections()
				return nil
			default:
				logger.Logerr("Error connecting: " + err.Error())
				return err
			}
		}
		logger.Loginfo("Client connected.")
		logger.Loginfo("Client " + c.RemoteAddr().String() + " connected.")

		go connectionLoop(c, &logger)
	}
}

// Warning! We need an intelligent system for closing the socket when we panic.
// likely a `defer DeleteHost(host)`?
func StartHostConnection(port string) (TipTapHost, error) {
	host := makeEmptyHost()
	logger := host.logger
	l, err := net.Listen(CONNTYPE, combineIPAndPort(CONNHOST, port))
	if err != nil {
		logger.Logerr("Failed to create host socket: " + err.Error())
		return host, err
	}

	host.listener = l
	return host, nil
}

func StopHostConnection(host TipTapHost) bool {
	logger := host.logger
	err := host.listener.Close()
	if err != nil {
		logger.Logerr("Failed to delete host socket: " + err.Error())
		return false
	}
	return true
}
