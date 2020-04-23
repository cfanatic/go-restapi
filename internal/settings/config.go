package settings

import (
	"flag"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type keys int

const (
	PATH        = "cmd/netchat/config.toml"
	PATH_D      = "config.toml"
	PORT   keys = iota
)

var (
	mode = flag.String("mode", PATH, "define execution mode")
)

type config struct {
	General general
}

type general struct {
	Port int
}

func Get(key keys) interface{} {
	var (
		conf config
		path string
	)
	if flag.Parse(); *mode == "debug" {
		path, _ = filepath.Abs(PATH_D)
	} else {
		path, _ = filepath.Abs(*mode)
	}
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		panic(err)
	}
	switch key {
	case PORT:
		return conf.General.Port
	default:
		return nil
	}
}

func Port() int {
	return Get(PORT).(int)
}
