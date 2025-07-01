package store

import (
	"encoding/json"
	"log"
	"os"
)

type Task struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

func Save(lines map[string]Task) {
	file, err := os.Create("tmp/dat2.json")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(lines); err != nil {
		log.Fatalf("failed to encode data: %v", err)
	}
}

func Load() map[string]Task {
	jsonData, _ := os.ReadFile("tmp/dat2.json")

	tasks := make(map[string]Task)

	err := json.Unmarshal(jsonData, &tasks)
	if err != nil {
		log.Fatalf("Unable to marshal JSON due to %s", err)
	}
	return tasks
}
