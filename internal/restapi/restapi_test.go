package restapi

import (
	"testing"
)

func TestSendRequest(t *testing.T) {
	INPUT := "POST called"
	request := Request{
		Token:  "fake_token",
		Method: "POST",
		Url:    "https://127.0.0.1",
		Body:   "Send POST request",
	}
	message := SendRequest(request)
	if message != INPUT {
		t.Errorf("New() failed -> want: \"%s\", got: \"%s\"", INPUT, message)
	}
}
