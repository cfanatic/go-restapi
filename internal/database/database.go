package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cfanatic/go-netchat/internal/settings"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type ttime = time.Time

type Database struct {
	db   *sql.DB
	cred *[]Credential
}

type Message struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Date       ttime  `json:"date"`
	Salt       string `json:"salt"`
	Message    string `json:"msg"`
	ReadL      int    `json:"read_local"`
	ReadR      int    `json:"read_remote"`
	Auxiliary  int    `json:"aux"`
	Encryption int    `json:"enc"`
}

type Credential struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

var (
	configM settings.Mysql
	configT settings.Token
)

func New() (*Database, error) {
	var db Database
	var err error
	dataSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		configM.GetUser(),
		configM.GetPassword(),
		configM.GetAddress(),
		configM.GetPort(),
		configM.GetDatabase(),
	)
	to := configM.GetTimeout()
	db.db, err = sql.Open("mysql", dataSource)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*to)
	defer cancel()
	if err = db.db.PingContext(ctx); err != nil {
		err = errors.New("Could not connect to database server")
	}
	return &db, err
}

func (db *Database) GetUser(user string) (*Credential, error) {
	var err error
	query := &(sql.Rows{})
	cred := Credential{}
	if query, err = db.db.Query("SELECT * FROM users WHERE user=?", user); err == nil {
		for query.Next() {
			err = query.Scan(&cred.ID, &cred.Name, &cred.User, &cred.Password)
		}
		if cred == (Credential{}) {
			err = errors.New(fmt.Sprintf(`User "%s" is not available in the database`, user))
			return &cred, err
		}
	}
	return &cred, err
}

func (db *Database) UpdatePassword(user, password string) error {
	var (
		hash  []byte
		query *sql.Rows
		res   sql.Result
		err   error
	)
	checksum := GenerateHash(password)
	salt := configT.GetSecretKey()
	tmp := []byte(checksum + salt)
	if hash, err = bcrypt.GenerateFromPassword(tmp, bcrypt.DefaultCost); err == nil {
		if query, err = db.db.Query("SELECT EXISTS(SELECT 1 FROM users WHERE user=?)", user); err == nil {
			var cnt int
			query.Next()
			query.Scan(&cnt)
			if cnt == 0 {
				err = errors.New("User is not available in database: " + user)
				return err
			}
		}
		if res, err = db.db.Exec("UPDATE users SET password=? WHERE user=?", hash, user); err == nil {
			if cnt, err := res.RowsAffected(); err == nil {
				if cnt != 1 {
					err = errors.New("Could not update user password")
					return err
				}
			}
		}
	}
	return err
}

func GenerateHash(password string) string {
	salt := configT.GetSecretKey()
	sum := sha256.New()
	sum.Write([]byte(password + salt))
	return fmt.Sprintf("%x", sum.Sum(nil))
}
