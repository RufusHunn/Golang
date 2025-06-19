package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// func save(data map[string][]byte) {
func save(data []string) {
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

func load() []string {
	file, err := os.Open("tmp/dat1.json")
	if err != nil {
		log.Fatalf("failed to find file: %v", err)
	}
	defer file.Close()

	var contents []string
	json.NewDecoder(file).Decode(&contents)
	fmt.Println("Found contents : ", contents)
	return contents
}

// Run below like:
// go build main.go
// ./main.exe --item='todo list' --status=completed

func main() {

	item := flag.String("item", "activity", "Your activity")
	status := flag.String("status", "not completed", "Item status")

	flag.Parse()

	fmt.Println("Item: ", *item)
	fmt.Println("Status: ", *status)

	lines := append(load(), *item)

	fmt.Println("Preparing to write lines: ", lines)

	save(lines)
}
