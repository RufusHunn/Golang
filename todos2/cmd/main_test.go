package main

import (
	"testing"
	"todos2/store"
)

// Below test to confirm individual pass
func TestIndividual(t *testing.T) {
	Tasks := make(map[string]store.Task)
	TaskChan := make(chan TaskMessage)
	payload := store.Task{Description: "myTest", Status: "todo"}
	go taskActor(TaskChan, Tasks)
	TaskChan <- TaskMessage{Action: "create", Key: "1", Payload: payload, Response: make(chan interface{})}
	task, exists := Tasks["1"]
	if !exists || task.Description != "myTest" || task.Status != "todo" {
		t.Errorf("Create - Lines content not as expected")
	}
}

// Using Parallel to validate concurrent-safe
func TestConcurrentFunctions(t *testing.T) {
	Tasks := make(map[string]store.Task)
	TaskChan := make(chan TaskMessage)
	t.Run("SubtestActorCreate", func(t *testing.T) {
		t.Parallel()
		payload := store.Task{Description: "first", Status: "todo"}
		go taskActor(TaskChan, Tasks)
		TaskChan <- TaskMessage{Action: "create", Key: "1", Payload: payload, Response: make(chan interface{})}
		task, exists := Tasks["1"]
		if !exists || task.Description != "first" || task.Status != "todo" {
			t.Errorf("Create - Lines content not as expected")
		}
	})
	t.Run("SubtestActorRead", func(t *testing.T) {
		t.Parallel()
		Tasks["2"] = store.Task{Description: "second", Status: "todo"}
		go taskActor(TaskChan, Tasks)
		TaskChan <- TaskMessage{Action: "read", Key: "1", Response: make(chan interface{})}
		task, exists := Tasks["2"]
		if !exists || task.Description != "second" || task.Status != "todo" {
			t.Errorf("Read - Unexpected line content")
		}
	})
	t.Run("SubtestActorUpdate", func(t *testing.T) {
		t.Parallel()
		Tasks["3"] = store.Task{Description: "third", Status: "todo"}
		payload := store.Task{Description: "third", Status: "completed"}
		go taskActor(TaskChan, Tasks)
		TaskChan <- TaskMessage{Action: "update", Key: "3", Payload: payload, Response: make(chan interface{})}
		updatedTask, exists := Tasks["3"]
		if !exists || updatedTask.Description != "third" || updatedTask.Status != "completed" {
			t.Errorf("Update - Unexpected line content")
		}

	})
	t.Run("SubtestActorDelete", func(t *testing.T) {
		t.Parallel()
		Tasks["4"] = store.Task{Description: "fourth", Status: "todo"}
		go taskActor(TaskChan, Tasks)
		TaskChan <- TaskMessage{Action: "delete", Key: "4", Response: make(chan interface{})}
		_, exists := Tasks["4"]
		if exists {
			t.Errorf("Delete - Deleted line still exists")
		}
	})
}
