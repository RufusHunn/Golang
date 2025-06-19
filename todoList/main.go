package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
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

	f, err := os.Open("tmp/dat1")
	check(err)
	defer f.Close()

	b1 := make([]byte, 32)
	m1, err := f.Read(b1)
	check(err)
	readFileAsString := string(b1)

	fmt.Println("Existing lines: ", readFileAsString, " with len ", m1)

	lines := []string{readFileAsString, *item}
	output := strings.Join(lines, "\n")
	d1 := []byte(output)
	fmt.Println("Preparing to write lines: ", string(d1))

	// erx := os.WriteFile("tmp/dat1", d1, 0644) // This does not actually write - why?
	erx := os.WriteFile("tmp/dat1", []byte(*item), 0644) // Writes latest input but we want all lines!
	check(erx)
	fmt.Println("wrote lines: ", d1)

	f.Sync()

}
