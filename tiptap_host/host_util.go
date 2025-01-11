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

func readConnection(conn net.Conn, logger *Logger) bool {
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		logger.Logerr("Client left.")
		conn.Close()
		return true
	}

	logger.Loginfo("Client message: " + string(buffer[:len(buffer)-1]))
	conn.Write(buffer)
	return false
}

func loopMetrics(conn net.Conn) {
	// TODO: if we want to store info like "you've sent X messages so far" or whatever
}

func combineIPAndPort(ip, port string) string {
	return ip + ":" + port
}
