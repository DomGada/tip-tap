package shared

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
)

func GetChoice(list []string) string {
	fmt.Println("Choose from the following:")
	for idx := range list {
		line := list[idx]
		fmt.Printf("%d - %s\n", idx+1, line)
	}
	var chosen int
	for {
		var err error
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter a number: ")
		buff, _ := reader.ReadString('\n')
		buff = buff[:len(buff)-1]

		chosen, err = strconv.Atoi(buff)

		if err == nil && (chosen >= 1 && chosen <= len(list)) {
			break
		}
	}
	return list[chosen-1]
}

func GetPort() string {
	var port string
	for {
		var tmpconv int
		var err error
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter a number: ")
		port, _ = reader.ReadString('\n')
		port = port[:len(port)-1]

		tmpconv, err = strconv.Atoi(port)

		if err == nil && (tmpconv >= 1 && tmpconv <= int(math.Pow(2, 16))) {
			break
		}
	}
	return port

}
