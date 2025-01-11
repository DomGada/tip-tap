package tiptaphost

import (
	"bufio"
	"net"
)

func makeEmptyHost() TipTapHost {
	return TipTapHost{nil, Logger{}}
}

func connectionLoop(conn net.Conn, logger *Logger) {
	exited := false
	for exited == false {
		exited = readConnection(conn, logger)
		loopMetrics(conn)
	}
}
// We need to store a record of what the client sent.
// When we enter our read connection we want to verify
// that this is indeed what our client.
// Do we need that validation?
func readConnection(conn net.Conn, logger *Logger) bool {
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		logger.Logerr("Client left.")
		conn.Close()
		return true
	}

	logger.Loginfo("Client message: " + string(buffer[:len(buffer)-1]))
	// TODO: Replace below with go routine of seperate connection itteration function.
	conn.Write(buffer)
	return false
}

func (conns []net.Conn, msgs []string ){
	// TODO: for all the fucking connections, YEAT THE THE MESSAGE 
	// Yeat the msgs bro

	// PSUEDO
	// FOR CONN IN CONS POP QUEUE SEND SAME MESSAGE TO ALL CONS, and MOVE DOWN THE QUEUE

}
func loopMetrics(conn net.Conn) {
	// TODO: if we want to store info like "you've sent X messages so far" or whatever
}

func combineIPAndPort(ip, port string) string {
	return ip + ":" + port
}
