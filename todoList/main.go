package main

import (
	"flag"
	"fmt"
)

// Run below like:
// go build main.go
// ./main.exe --item=todo --status=completed

func main() {
	item := flag.String("item", "activity", "Your activity")
	status := flag.String("status", "not completed", "Item status")

	flag.Parse()

	fmt.Println("Item: ", *item)
	fmt.Println("Status: ", *status)
}
