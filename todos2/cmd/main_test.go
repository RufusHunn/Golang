package main

import (
	"testing"
	"todos2/store"
)

// Below test to confirm individual pass
func TestIndividual(t *testing.T) {
	payload := store.Task{Description: "myTest", Status: "todo"}
	go taskActor()
	TaskChan <- TaskMessage{Action: "create", Key: "1", Payload: payload, Response: make(chan interface{})}
	task, exists := Tasks["1"]
	if !exists || task.Description != "myTest" || task.Status != "todo" {
		t.Errorf("Create - Lines content not as expected")
	}
}

func TestGenerateID_FirstVal(t *testing.T) {
	Tasks = make(map[string]store.Task)
	output := generateID()
	if output != "1" {
		t.Errorf("GenerateID - First val not 1 as expected but was %s", output)
	}
}

func TestGenerateID_NextVal(t *testing.T) {
	Tasks = make(map[string]store.Task)
	Tasks["1"] = store.Task{Description: "dummy", Status: "todo"}
	output := generateID()
	if output != "2" {
		t.Errorf("GenerateID - Incremented val not 2 as expected but was %s", output)
	}
}

// Using Parallel to validate concurrent-safe
// NOTE: Unfortunately I was not able to get test to pass in time and confirm solution as concurrent safe.
// If I had more time I'd test with httptest library and make endpoint calls rather than testing taskActor alone
func TestConcurrentFunctions(t *testing.T) {
	go taskActor()
	t.Run("SubtestActorCreate", func(t *testing.T) {
		t.Parallel()
		payload := store.Task{Description: "first", Status: "todo"}
		TaskChan <- TaskMessage{Action: "create", Key: "1", Payload: payload, Response: make(chan interface{})}
		task, exists := Tasks["1"]
		if !exists || task.Description != "first" || task.Status != "todo" {
			t.Errorf("Create - Lines content not as expected, found: %t %s %s Len is: %d", exists, task.Description, task.Status, len(Tasks))
		}
	})
	t.Run("SubtestActorRead", func(t *testing.T) {
		t.Parallel()
		Tasks["2"] = store.Task{Description: "second", Status: "todo"}
		TaskChan <- TaskMessage{Action: "read", Key: "2", Response: make(chan interface{})}
		task, exists := Tasks["2"]
		if !exists || task.Description != "second" || task.Status != "todo" {
			t.Errorf("Read - Unexpected line content")
		}
	})
	t.Run("SubtestActorUpdate", func(t *testing.T) {
		t.Parallel()
		Tasks["3"] = store.Task{Description: "third", Status: "todo"}
		payload := store.Task{Description: "third", Status: "completed"}
		TaskChan <- TaskMessage{Action: "update", Key: "3", Payload: payload, Response: make(chan interface{})}
		updatedTask, exists := Tasks["3"]
		if !exists || updatedTask.Description != "third" || updatedTask.Status != "completed" {
			t.Errorf("Update - Unexpected line content")
		}

	})
	t.Run("SubtestActorDelete", func(t *testing.T) {
		t.Parallel()
		Tasks["4"] = store.Task{Description: "fourth", Status: "todo"}
		TaskChan <- TaskMessage{Action: "delete", Key: "4", Response: make(chan interface{})}
		_, exists := Tasks["4"]
		if exists {
			t.Errorf("Delete - Deleted line still exists")
		}
	})
}
