package store

import (
	"testing"
)

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
