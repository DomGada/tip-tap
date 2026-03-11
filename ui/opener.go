package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/DomGada/Tip-Tap/client"
	"github.com/DomGada/Tip-Tap/discovery"
	"github.com/DomGada/Tip-Tap/host"
	"github.com/DomGada/Tip-Tap/shared"
)

const (
	CONNECT = 0
	SEARCH  = 1
	HOST    = 2
)

var opener_list = []string{
	"Connect to chat",
	"Search for chat",
	"Host your own chat",
	"Exit",
}

func Opener() int {
	choice := shared.GetChoice(opener_list)

	switch choice {
	case "Exit":
		return 0
	case opener_list[CONNECT]:
		return promptAndRunClient()
	case opener_list[SEARCH]:
		return promptAndSearch()
	case opener_list[HOST]:
		return promptAndRunHost()
	default:
		err := fmt.Errorf(
			"%d Somehow you entered a scenario that should be unreachable. Congratulations!\n",
			shared.ERR_OPENER_BAD_CHOICE,
		)
		panic(err)

	}
}

func promptAndRunClient() int {
	fmt.Println("Connecting to TipTap chat...")

	ip := getIPAddress()
	port := shared.GetPort()
	err := client.ConnectClient(ip, port)
	if err != nil {
		fmt.Printf("Failed to connect: %s\n", err.Error())
		return 1
	}
	return 0
}

func getIPAddress() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter host IP address: ")
		input, _ := reader.ReadString('\n')
		input = input[:len(input)-1]
		if len(input) > 0 {
			return input
		}
	}
}

func promptAndSearch() int {
	fmt.Println("Searching for TipTap hosts on the local network...")
	fmt.Println("Listening for 5 seconds...")

	hosts, err := discovery.SearchHosts()
	if err != nil {
		fmt.Printf("Search failed: %s\n", err.Error())
		return 1
	}

	if len(hosts) == 0 {
		fmt.Println("No hosts found on the local network.")
		fmt.Println("Use 'Connect to chat' to connect manually.")
		return 0
	}

	choices := make([]string, len(hosts))
	for i, h := range hosts {
		choices[i] = fmt.Sprintf("%s (%s:%s)", h.Name, h.IP, h.TCPPort)
	}
	choices = append(choices, "Cancel")

	selected := shared.GetChoice(choices)
	if selected == "Cancel" {
		return 0
	}

	for _, h := range hosts {
		label := fmt.Sprintf("%s (%s:%s)", h.Name, h.IP, h.TCPPort)
		if label == selected {
			fmt.Printf("Connecting to %s...\n", h.Name)
			err := client.ConnectClient(h.IP, h.TCPPort)
			if err != nil {
				fmt.Printf("Failed to connect: %s\n", err.Error())
				return 1
			}
			return 0
		}
	}

	return 0
}

func promptAndRunHost() int {
	fmt.Println("Starting TipTap Host...")

	port := shared.GetPort()
	err := host.RunHost(context.Background(), port)
	if err != nil {
		fmt.Printf("Failed to run host: %s\n", err.Error())
		return 1
	}
	fmt.Println("TipTap Host is running.")
	return 0
}
