package store

import (
	"encoding/json"
	"log"
	"os"
)

func Save(data map[string]string) {
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

func Load() map[string]string {
	file, err := os.Open("tmp/dat1.json")
	if err != nil {
		log.Fatalf("failed to find file: %v", err)
	}
	defer file.Close()

	var contents map[string]string
	json.NewDecoder(file).Decode(&contents)
	log.Println("Found file contents : ", contents)
	return contents
}

func ApplyFlags(lines map[string]string, item, status, todelete string) map[string]string {
	_, found := lines[todelete]
	if found {
		delete(lines, todelete)
	}

	if item != "" {
		lines[item] = status
	}

	log.Println("Updated lines: ", lines)

	return lines
}
