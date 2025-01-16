package tiptaphost

/*
This defines our public library functions.
See host_util.go for utility functions
*/

import (
	"net"
	"fmt"
)

const (
	CONNHOST = "localhost"
	CONNTYPE = "tcp"
) // TODO: we will want to make this more functional lol



func RunHost(port string) error {
	var err error // this is not needed due to the next line, but is here for clarity
	hostConn, err := StartHostConnection(port)
	logger := hostConn.logger
	if err != nil {
		logger.Logerr("Unable to form connection with host, " + err.Error())
		return err
	}
	hostSock := hostConn.listener
	for {
		fmt.Printf("Waiting for additional clients to connect...\n")
		c, err := hostSock.Accept()
		if err != nil {
			logger.Logerr("Error connecting: " + err.Error())
			break
		}
		logger.Loginfo("Client connected.")
		logger.Loginfo("Client " + c.RemoteAddr().String() + " connected.")

		go connectionLoop(c, &logger)
	}

	return err
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
