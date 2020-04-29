package logger

import (
	"io"
	"log"
	"os"

	"github.com/cfanatic/go-netchat/internal/mode"
)

var (
	Log *log.Logger
)

func init() {
	var (
		path string
		file *os.File
		err  error
	)
	path = mode.GetFilePath("netchat.log")
	if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err == nil {
		Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
		mwriter := io.MultiWriter(os.Stdout, file)
		Log.SetOutput(mwriter)
	} else {
		panic(err)
	}
}
