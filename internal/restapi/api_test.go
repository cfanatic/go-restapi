package restapi

import "testing"

func TestNew(t *testing.T) {
	INPUT := "restapi"
	v := New()
	if v != INPUT {
		t.Errorf("New() failed -> want: \"%s\", got: \"%s\"", INPUT, v)
	}
}
