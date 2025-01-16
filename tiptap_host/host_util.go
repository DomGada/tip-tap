package tiptaphost

import (
	"bufio"
	"net"
	"sync"
)
var (
	connections []net.Conn
	connMutex sync.Mutex
)

func makeEmptyHost() TipTapHost {
	return TipTapHost{nil, Logger{}}
}

func connectionLoop(conn net.Conn, logger *Logger) {
	// Add our new conn to list
	addConnection(conn)
	defer removeConnection(conn)
	sendDirectMessage(conn, "Welcome to the TipTap Host!")
	sendDirectMessage(conn, "Please enter your display name :3")

	username, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		logger.Logerr("Error reading username: " + err.Error())
		conn.Close()
		return
	}
	logger.Loginfo("Client username: " + username)
	exited := false
	for exited == false {
		exited = readConnection(conn, logger, username)
		loopMetrics(conn)
	}
}

func readConnection(conn net.Conn, logger *Logger,username string) bool {
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		logger.Logerr(username + " left.")
		conn.Close()
		return true
	}

	userMessage :=  username + ": " + string(buffer[:len(buffer)-1])

	logger.Loginfo(userMessage)
	broadcastMessage(conn,userMessage)
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
		if c == conn{
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	conn.Close()
}

func broadcastMessage(sender net.Conn, msg string){
	connMutex.Lock()
	defer connMutex.Unlock()
	
	for _, c := range connections{
		if c != sender{
			sendDirectMessage(c,msg)
		}
	}
}

func sendDirectMessage(conn net.Conn, msg string) {
	conn.Write([]byte(msg + "\n"))
}

func loopMetrics(conn net.Conn) {
	// TODO: if we want to store info like "you've sent X messages so far" or whatever
}

func combineIPAndPort(ip, port string) string {
	return ip + ":" + port
}
