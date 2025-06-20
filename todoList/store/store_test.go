package store

import (
	"testing"
)

func TestApplyFlagsCreate(t *testing.T) {

	lines := map[string]string{"first": "completed"}

	// expectedLines = map[string]string {"first": "completed", "second": "in progress"}

	var item string = "second"
	var status string = "in progress"

	editedLines := ApplyFlags(lines, item, status, "")
	if len(editedLines) != 2 {
		t.Errorf("applyFlags = %d; expected len 2", len(editedLines))
	}
}

func TestApplyFlagsDelete(t *testing.T) {

	lines := map[string]string{"to remove": "completed"}

	// expectedLines = map[string]string {}

	var todelete string = "to remove"

	editedLines := ApplyFlags(lines, "", "", todelete)
	if len(editedLines) != 0 {
		t.Errorf("applyFlags = %d; expected empty", len(editedLines))
	}
}
