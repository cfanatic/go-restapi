package restapi

import (
	"fmt"
	"testing"
	"time"

	"github.com/cfanatic/go-netchat/internal/settings"
)

var configG settings.General

func TestSendRequest(t *testing.T) {
	cred := configG.GetTestUser()
	SendRequest(Request{
		Method: "GET",
		Url:    fmt.Sprintf("https://127.0.0.1:1025/login/%s/%s", cred[0], cred[1]),
		Message: Message{
			Name: cred[0],
			Date: time.Now().Format("2006-01-02 15:04:05"),
			Text: "This is a test",
		},
	})
}
