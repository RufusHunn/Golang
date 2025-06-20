package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"todoList/store"
)

func main() {

	item := flag.String("item", "", "Your activity")
	status := flag.String("status", "", "Item status")
	todelete := flag.String("delete", "", "Description to delete")

	flag.Parse()

	fmt.Println("Item: ", *item)
	fmt.Println("Status: ", *status)
	fmt.Println("Delete: ", *todelete)

	lines := store.Load()

	editedLines := store.ApplyFlags(lines, *item, *status, *todelete)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	fmt.Println("App is running. Press Ctrl+C to exit...")
	<-sigChan

	store.Save(editedLines)
}
