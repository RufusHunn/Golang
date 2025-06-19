package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func save(data map[string]string) {
	file, err := os.Create("tmp/dat1.json")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("failed to encode data: %v", err)
	}
}

func load() map[string]string {
	file, err := os.Open("tmp/dat1.json")
	if err != nil {
		log.Fatalf("failed to find file: %v", err)
	}
	defer file.Close()

	var contents map[string]string
	json.NewDecoder(file).Decode(&contents)
	fmt.Println("Found contents : ", contents)
	return contents
}

// Run below like:
// go build main.go
// ./main.exe --item='todo list' --status=completed

func main() {

	item := flag.String("item", "", "Your activity")
	status := flag.String("status", "", "Item status")
	todelete := flag.String("delete", "", "Description to delete")

	flag.Parse()

	fmt.Println("Item: ", *item)
	fmt.Println("Status: ", *status)
	fmt.Println("Delete: ", *todelete)

	lines := load()

	_, found := lines[*todelete]
	if found {
		delete(lines, *todelete)
	}

	if *item != "" {
		lines[*item] = *status
	}

	fmt.Println("Preparing to write lines: ", lines)

	save(lines)
}
