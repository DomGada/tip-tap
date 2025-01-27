package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	CONNHOST = "localhost"
	CONNTYPE = "tcp"
)

func ConnectClient(ip_addr, port string) error {

	var err error // this is not needed due to the next line, but is here for clarity
	client, err := startClientConnection(ip_addr, port)
	logger := client.logger
	if err != nil {
		logger.Logerr("Unable to form connection with host, " + err.Error())
		return err
	}

	sock := client.connection
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Text to send: ")
		input, _ := reader.ReadString('\n')
		sock.Write([]byte(input))
		message, _ := bufio.NewReader(sock).ReadString('\n')
		logger.Loginfo("Server relay: " + message)
	}
}

func startClientConnection(ip_addr, port string) (TipTapClient, error) {
	client := makeEmptyClient()
	logger := client.logger
	conn, err := net.Dial(CONNTYPE, ip_addr+":"+port)
	if err != nil {
		logger.Logerr("Failed to connect: " + err.Error())
		return client, err
	}

	client.connection = conn
	return client, nil

}
