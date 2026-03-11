package host

import (
	"bufio"
	"net"
	"sync"

	"github.com/DomGada/Tip-Tap/shared"
)

var (
	connections []net.Conn
	connMutex   sync.Mutex
)

func makeEmptyHost() TipTapHost {
	return TipTapHost{nil, shared.Logger{}}
}

func connectionLoop(conn net.Conn, logger *shared.Logger) {
	// Add our new conn to list
	addConnection(conn)
	defer removeConnection(conn)
	sendDirectMessageLine(conn, "Welcome to the TipTap Host!")
	sendDirectMessageLine(conn, "Please enter your display name :3")

	reader := bufio.NewReader(conn)
	username, err := reader.ReadString('\n')
	if err != nil {
		logger.Logerr("Error reading username: " + err.Error())
		conn.Close()
		return
	}
	username = username[:len(username)-1] // trim trailing newline
	logger.Loginfo("Client username: " + username)
	broadcastMessage(conn, "** "+username+" has joined the chat **")
	exited := false
	for !exited {
		exited = readConnection(conn, reader, logger, username)
		loopMetrics(conn)
	}
	broadcastMessage(conn, "** "+username+" has left the chat **")
}

func readConnection(conn net.Conn, reader *bufio.Reader, logger *shared.Logger, username string) bool {
	buffer, err := reader.ReadBytes('\n')
	if err != nil {
		logger.Logerr(username + " left.")
		conn.Close()
		return true
	}

	userMessage := username + ": " + string(buffer[:len(buffer)-1])

	logger.Loginfo(userMessage)
	broadcastMessage(conn, userMessage)
	return false
}

func addConnection(conn net.Conn) {
	connMutex.Lock()
	connections = append(connections, conn)
	connMutex.Unlock()
}

func removeConnection(conn net.Conn) {
	connMutex.Lock()
	defer connMutex.Unlock()
	for i, c := range connections {
		if c == conn {
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	conn.Close()
}

func closeAllConnections() {
	connMutex.Lock()
	defer connMutex.Unlock()
	for _, c := range connections {
		sendDirectMessageLine(c, "** Server is shutting down **")
		c.Close()
	}
	connections = nil
}

func broadcastMessage(sender net.Conn, msg string) {
	connMutex.Lock()
	defer connMutex.Unlock()

	for _, c := range connections {
		if c != sender {
			sendDirectMessageLine(c, msg)
		}
	}
}

func sendDirectMessage(conn net.Conn, msg string) {
	conn.Write([]byte(msg))
}

func sendDirectMessageLine(conn net.Conn, msg string) {
	sendDirectMessage(conn, msg+"\n")
}

func loopMetrics(conn net.Conn) {
	// TODO: if we want to store info like "you've sent X messages so far" or whatever
}

func combineIPAndPort(ip, port string) string {
	return ip + ":" + port
}
