package tiptaphost

/*
This defines our public library functions.
See host_util.go for utility functions
*/

import (
	"log"
	"net"
	"os"
)

const (
	CONNHOST = "localhost"
	CONNTYPE = "tcp"
) // TODO: we will want to make this more functional lol

type TipTapHostSocket net.Listener

// Warning! We need an intelligent system for closing the socket when we panic.
// likely a `defer DeleteHost(host)`?
func CreateHost(port string) (TipTapHostSocket, error) {
	l, err := net.Listen(CONNTYPE, combineIPAndPort(CONNHOST, port))
	if err != nil {
		logger := log.Default()
		logger.SetOutput(os.Stderr)
		logger.Println("Failed to create host socket: " + err.Error())
		return nil, err
	}
	return l, nil
}

func DeleteHost(host TipTapHostSocket) bool {
	err := host.Close()
	if err != nil {
		logger := log.Default()
		logger.SetOutput(os.Stderr)
		logger.Println("Failed to delete host socket: " + err.Error())
		return false
	}
	return true
}

func ReadHost(host TipTapHostSocket) (string, bool) {
	return "", false
}
