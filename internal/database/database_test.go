package database

import (
	"fmt"
	"testing"
	"time"

	Log "github.com/cfanatic/go-netchat/internal/logger"
)

func TestNew(t *testing.T) {
	if db, err := New(); err == nil {
		if tmp, err := db.GetMessages(0, 3); err == nil {
			for _, val := range *tmp {
				fmt.Println(val.ID, "     ", val.Name, "     ", string(val.Date), "     ", val.Message)
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
			Date:    []byte(time.Now().Format("2006-01-02 15:04:05")),
			Salt:    "n/a",
			Message: "This is a test",
		}
		fmt.Println(db.SendMessage(msg))
	}
}
