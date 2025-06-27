package store

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

type Task struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

var Lines map[string]Task

func Save() {
	file, err := os.Create("tmp/dat2.json")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(Lines); err != nil {
		log.Fatalf("failed to encode data: %v", err)
	}
}

func Load() error {
	jsonData, _ := os.ReadFile("tmp/dat2.json")

	tasks := make(map[string]Task)

	err := json.Unmarshal(jsonData, &tasks)
	if err != nil {
		log.Fatalf("Unable to marshal JSON due to %s", err)
	}
	Lines = tasks
	return nil
}

func Create(description string, status string) {

	keys := make([]int, 0, len(Lines))
	for k := range Lines {
		ik, _ := strconv.Atoi(k)
		keys = append(keys, ik)
	}
	nextKey := keys[len(keys)-1] + 1
	sk := strconv.Itoa(nextKey)
	Lines[sk] = Task{Description: description, Status: status}
}

func Update(ix, description, status string) {
	Lines[ix] = Task{Description: description, Status: status}
}

func Delete(ix string) {
	_, found := Lines[ix]
	if found {
		delete(Lines, ix)
	}
}

func List() map[string]Task {
	return Lines
}

func Get(index string) Task {
	return Lines[index]
}
