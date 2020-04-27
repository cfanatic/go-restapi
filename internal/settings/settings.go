package settings

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

type General struct {
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
	Token   Token
	Mysql   Mysql
}

const (
	PATH_RELEASE = "cmd/netchat/config.toml"
	PATH_DEBUG   = "../../misc/config.toml"
)

var (
	mode   = flag.String("mode", PATH_RELEASE, "define release mode")
	config = func() Settings {
		var (
			config Settings
			path   string
			e      *os.PathError
		)
		if flag.Parse(); *mode == "debug" {
			path, _ = filepath.Abs(PATH_DEBUG)
		} else {
			path, _ = filepath.Abs(*mode)
		}
		if _, err := toml.DecodeFile(path, &config); err != nil {
			if errors.As(err, &e) {
				log.Println("Using default configuration setting")
				config = Settings{
					General{
						Address:     "127.0.0.1",
						Port:        8080,
						Port_TLS:    443,
						Certificate: []string{"misc/server.crt", "misc/server.key"},
					},
					Token{},
					Mysql{},
				}
			} else {
				log.Println(err)
			}
		}
		return config
	}()
)

func (general General) GetAddress() string {
	return config.General.Address
}

func (general General) GetPort() int {
	return config.General.Port
}
func (general General) GetPortTLS() int {
	return config.General.Port_TLS
}
func (general General) GetCertificate() []string {
	cert := config.General.Certificate
	path_crt, _ := filepath.Abs(cert[0])
	path_key, _ := filepath.Abs(cert[1])
	return []string{path_crt, path_key}
}

func (token Token) GetSecretKey() string {
	return config.Token.SecretKey
}
func (token Token) GetExpiration() time.Duration {
	return time.Duration(config.Token.Expiration)
}

func (mysql Mysql) GetUser() string {
	return config.Mysql.User
}

func (mysql Mysql) GetPassword() string {
	return config.Mysql.Password
}

func (mysql Mysql) GetAddress() string {
	return config.Mysql.Address
}

func (mysql Mysql) GetPort() int {
	return config.Mysql.Port
}

func (mysql Mysql) GetDatabase() string {
	return config.Mysql.Database
}

func (mysql Mysql) GetPeer() string {
	return config.Mysql.Peer
}
