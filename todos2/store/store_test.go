package store

import (
	"testing"
)

func TestCreate(t *testing.T) {

	Lines = make(map[string]Task)

	Create("first", "started")
	if len(Lines) != 1 {
		t.Errorf("Create - Lines length = %d; expected one", len(Lines))
	}
}

func TestUpdate(t *testing.T) {

	Lines = map[string]Task{"1": {Description: "started", Status: "started"}}

	Update("1", "first", "completed")
	if len(Lines) != 1 || Lines["first"].Status != "completed" {
		t.Errorf("Create - Lines content changed unexpectedly")
	}
}

func TestGet(t *testing.T) {

	Lines = map[string]Task{"1": {Description: "started", Status: "completed"}}

	var found = Get("1")
	if found.Description != "started" || found.Status != "Completed" {
		t.Errorf("Get = 1; unexpected output: %s", found.Description)
	}
}

func TestDelete(t *testing.T) {

	Lines = map[string]Task{"1": {Description: "started", Status: "completed"}}

	Delete("first")
	if len(Lines) > 0 {
		t.Errorf("Delete = Lines length %d; expected empty", len(Lines))
	}
}
