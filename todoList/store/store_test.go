package store

import (
	"testing"
)

func TestCreate(t *testing.T) {

	Lines = make(map[string]string)

	Create("first", "started")
	if len(Lines) != 1 {
		t.Errorf("Create - Lines length = %d; expected one", len(Lines))
	}
}

func TestCreateItemExists(t *testing.T) {

	Lines = map[string]string{"first": "started"}

	var err = Create("first", "completed")
	if err.Error() != "item already exists" {
		t.Errorf("Create - Expected error for existing item")
	}
	if len(Lines) != 1 || Lines["first"] != "started" {
		t.Errorf("Create - Lines content changed unexpectedly")
	}
}

func TestUpdate(t *testing.T) {

	Lines = map[string]string{"first": "started"}

	Update("first", "completed")
	if len(Lines) != 1 || Lines["first"] != "completed" {
		t.Errorf("Create - Lines content changed unexpectedly")
	}
}

func TestUpdateItemNotExists(t *testing.T) {

	Lines = make(map[string]string)

	var err = Update("first", "completed")
	if err.Error() != "item does not exist" {
		t.Errorf("Create - Expected error for non-existing item")
	}
	if len(Lines) != 0 {
		t.Errorf("Create - Lines content changed unexpectedly")
	}
}

func TestGet(t *testing.T) {

	Lines = map[string]string{"first": "completed"}

	var found = Get("first")
	if len(found) != 1 {
		t.Errorf("Get = %d; expected one", len(found))
	}
}

func TestGetNotFound(t *testing.T) {

	Lines = map[string]string{"first": "completed"}

	var found = Get("second")
	if len(found) != 0 {
		t.Errorf("Get = %d; expected none", len(found))
	}
}

func TestDelete(t *testing.T) {

	Lines = map[string]string{"first": "completed"}

	Delete("first")
	if len(Lines) > 0 {
		t.Errorf("Delete = Lines length %d; expected empty", len(Lines))
	}
}

func TestDeleteNotFound(t *testing.T) {
	Lines = map[string]string{"first": "completed"}

	Delete("second")
	if len(Lines) != 1 {
		t.Errorf("Delete = Lines length %d; expected one only", len(Lines))
	}
}
