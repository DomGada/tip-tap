package ui

import (
	"fmt"

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

func promptAndRunClient() {
	panic("Unimplemented!")
}

func promptAndSearch() int {
	panic("Unimplemented!")
}

func promptAndRunHost() int {
	fmt.Println("Starting TipTap Host...")

	port := shared.GetPort()
	err := host.RunHost(port)
	if err != nil {
		fmt.Printf("Failed to run host: %s\n", err.Error())
		return 1
	}
	fmt.Println("TipTap Host is running.")
	return 0
}
