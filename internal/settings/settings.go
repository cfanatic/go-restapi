package settings

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	Log "github.com/cfanatic/go-netchat/internal/logger"
)

type General struct {
	LogPath string
}

type Backend struct {
	Address     string
	Port        int
	Port_TLS    int
	Certificate []string
}

type Token struct {
	SecretKey  string
	Expiration int
}

type Mysql struct {
	User     string
	Password string
	Address  string
	Port     int
	Database string
	Peer     string
}

type Settings struct {
	General General
	Backend Backend
	Token   Token
	Mysql   Mysql
}

const (
	PATH_TERMINAL = "misc/config.toml"
	PATH_DEBUG    = "../../misc/config.toml"
)

var (
	mode   = flag.String("mode", PATH_TERMINAL, "default configuration mode")
	config Settings
)

func init() {
	var (
		path string
		e    *os.PathError
	)
	if flag.Parse(); *mode == "terminal" {
		path, _ = filepath.Abs(PATH_TERMINAL)
	} else if *mode == "debug" {
		path, _ = filepath.Abs(PATH_DEBUG)
	} else {
		path, _ = filepath.Abs(*mode)
	}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		if errors.As(err, &e) {
			Log.Log.Println("Warning: Using default configuration setting")
			config = Settings{
				General{
					LogPath: "misc/netchat.log",
				},
				Backend{
					Address:     "127.0.0.1",
					Port:        8080,
					Port_TLS:    443,
					Certificate: []string{"misc/server.crt", "misc/server.key"},
				},
				Token{},
				Mysql{},
			}
		} else {
			Log.Log.Println(err)
		}
	}
	if *mode == "debug" {
		crt := &config.Backend.Certificate[0]
		key := &config.Backend.Certificate[1]
		*crt = "../../" + *crt
		*key = "../../" + *key
	}
}

func (_ General) GetLogPath() string {
	return config.General.LogPath
}

func (_ Backend) GetAddress() string {
	return config.Backend.Address
}

func (_ Backend) GetPort() int {
	return config.Backend.Port
}
func (_ Backend) GetPortTLS() int {
	return config.Backend.Port_TLS
}
func (_ Backend) GetCertificate() []string {
	cert := config.Backend.Certificate
	path_crt, _ := filepath.Abs(cert[0])
	path_key, _ := filepath.Abs(cert[1])
	return []string{path_crt, path_key}
}

func (_ Token) GetSecretKey() string {
	return config.Token.SecretKey
}
func (_ Token) GetExpiration() time.Duration {
	return time.Duration(config.Token.Expiration)
}

func (_ Mysql) GetUser() string {
	return config.Mysql.User
}

func (_ Mysql) GetPassword() string {
	return config.Mysql.Password
}

func (_ Mysql) GetAddress() string {
	return config.Mysql.Address
}

func (_ Mysql) GetPort() int {
	return config.Mysql.Port
}

func (_ Mysql) GetDatabase() string {
	return config.Mysql.Database
}

func (_ Mysql) GetPeer() string {
	return config.Mysql.Peer
}
