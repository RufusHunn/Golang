package store

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

var Lines map[string]string

func Save() {
	file, err := os.Create("tmp/dat1.json")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(Lines); err != nil {
		log.Fatalf("failed to encode data: %v", err)
	}
}

func Load() {
	file, err := os.Open("tmp/dat1.json")
	if err != nil {
		log.Fatalf("failed to find file: %v", err)
	}
	defer file.Close()

	var contents map[string]string
	json.NewDecoder(file).Decode(&contents)
	Lines = contents
}

func Create(item, status string) error {
	if Lines[item] != "" {
		return errors.New("item already exists")
	}
	if item != "" {
		Lines[item] = status
	}
	return nil
}
func Update(item, status string) error {
	if Lines[item] == "" {
		return errors.New("item does not exist")
	}
	if item != "" {
		Lines[item] = status
	}
	return nil
}

func Delete(item string) {
	_, found := Lines[item]
	if found {
		delete(Lines, item)
	}
}

func List() map[string]string {
	return Lines
}

func Get(item string) map[string]string {

	itemVal, found := Lines[item]
	if found {
		foundMap := make(map[string]string)
		foundMap[item] = itemVal
		return foundMap
	}
	return nil
}
