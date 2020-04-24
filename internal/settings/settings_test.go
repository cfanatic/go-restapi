package settings

import "testing"

func TestAddress(t *testing.T) {
	INPUT := "127.0.0.1"
	v := Address()
	if v != INPUT {
		t.Errorf("New() failed -> want: \"%s\", got: \"%s\"", INPUT, v)
	}
}

func TestPort(t *testing.T) {
	INPUT := 8080
	v := Port()
	if v != INPUT {
		t.Errorf("New() failed -> want: \"%d\", got: \"%d\"", INPUT, v)
	}
}

func TestPortTLS(t *testing.T) {
	INPUT := 443
	v := PortTLS()
	if v != INPUT {
		t.Errorf("New() failed -> want: \"%d\", got: \"%d\"", INPUT, v)
	}
}
