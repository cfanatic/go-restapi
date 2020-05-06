package database

import (
	"fmt"
	"testing"
	"time"

	Log "github.com/cfanatic/go-netchat/internal/logger"
)

func TestNew(t *testing.T) {
	if db, err := New(); err == nil {
		if tmp, err := db.GetMessages(0, 3, "<name>"); err == nil {
			for _, val := range *tmp {
				fmt.Println(val.ID, "     ", val.Name, "     ", val.Date, "     ", val.Message)
			}
		} else {
			Log.Log.Println(err)
		}
		if err := db.UpdatePassword("<user>", "<password>"); err != nil {
			Log.Log.Println(err)
		}
		fmt.Println(db.GetMessageCount())
		msg := Message{
			Name:    "cfanatic",
			Date:    time.Now(),
			Message: "This is a test",
		}
		fmt.Println(db.SendMessage(msg))
	}
}
