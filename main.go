package main

import (
    "fmt"
    "github.com/DomGada/Tip-Tap/tiptap_host"
)

func main() {
    fmt.Println("Starting TipTap Host...")

    port := "25565" // Specify the port you want to use
    err := tiptaphost.RunHost(port)
    if err != nil {
        fmt.Printf("Failed to run host: %s\n", err.Error())
    } else {
        fmt.Println("TipTap Host is running.")
    }
}