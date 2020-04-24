package settings

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type keys int

const (
	PATH_R       = "cmd/netchat/config.toml"
	PATH_D       = "../../misc/config.toml"
	ADDRESS keys = iota
	PORT
	PORT_TLS
)

var (
	mode = flag.String("mode", PATH_R, "define release mode")
)

type config struct {
	General general
}

type general struct {
	Address  string
	Port     int
	Port_TLS int
}

func new() config {
	return config{
		General: general{
			Address:  "127.0.0.1",
			Port:     8080,
			Port_TLS: 443,
		},
	}
}

func get(key keys) interface{} {
	var (
		conf config
		path string
		e    *os.PathError
	)
	if flag.Parse(); *mode == "debug" {
		path, _ = filepath.Abs(PATH_D)
	} else {
		path, _ = filepath.Abs(*mode)
	}
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		if errors.As(err, &e) {
			log.Println("using default configuration setting")
			conf = new()
		} else {
			panic(err)
		}
	}
	switch key {
	case ADDRESS:
		return conf.General.Address
	case PORT:
		return conf.General.Port
	case PORT_TLS:
		return conf.General.Port_TLS
	default:
		return nil
	}
}

func Address() string {
	return get(ADDRESS).(string)
}

func Port() int {
	return get(PORT).(int)
}

func PortTLS() int {
	return get(PORT_TLS).(int)
}
