package database

import "testing"

func TestNew(t *testing.T) {
	INPUT := "database"
	v := New()
	if v != INPUT {
		t.Errorf("New() failed -> want: \"%s\", got: \"%s\"", INPUT, v)
	}
}
