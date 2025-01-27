package main

import (
	"github.com/DomGada/Tip-Tap/ui"
	"os"
)

func main() {
	returncode := ui.Opener()
	os.Exit(returncode)
}
