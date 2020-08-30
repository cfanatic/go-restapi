package mode

import (
	"flag"
	"path/filepath"
)

const (
	PATH_INIT     = "misc/"
	PATH_TERMINAL = "misc/"
	PATH_DEBUG    = "../../misc/"
)

var mode = flag.String("mode", PATH_TERMINAL, "default configuration mode")

func GetFilePath(filename string) string {
	var path string
	if flag.Parse(); *mode == "init" {
		path, _ = filepath.Abs(PATH_INIT)
	} else if *mode == "terminal" {
		path, _ = filepath.Abs(PATH_TERMINAL)
	} else if *mode == "debug" {
		path, _ = filepath.Abs(PATH_DEBUG)
	} else {
		path, _ = filepath.Abs(*mode)
	}
	return path + "/" + filename
}

func GetMode() string {
	return *mode
}
