package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/cfanatic/go-netchat/internal/settings"
	_ "github.com/go-sql-driver/mysql"
)

type ttime = time.Time

type Database struct {
	db   *sql.DB
	cred *[]Credential
	err  error
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
	DatabaseTemp = map[string]string{
		"RandomUser1": "test1",
		"RandomUser2": "test2",
	}
	config settings.Mysql
)

func New() *Database {
	db := &Database{}
	dataSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		config.GetUser(),
		config.GetPassword(),
		config.GetAddress(),
		config.GetPort(),
		config.GetDatabase(),
	)
	if db.db, db.err = sql.Open("mysql", dataSource); db.err != nil {
		log.Println(db.err)
	}
	return db
}

func (db *Database) GetUsers() *[]Credential {
	sel := &(sql.Rows{})
	creds := []Credential{}
	if sel, db.err = db.db.Query("SELECT * FROM users ORDER BY id ASC"); db.err != nil {
		log.Println(db.err)
	} else {
		for sel.Next() {
			cred := Credential{}
			if db.err = sel.Scan(&cred.ID, &cred.Name, &cred.User, &cred.Password); db.err != nil {
				log.Println(db.err)
			}
			creds = append(creds, cred)
		}
	}
	return &creds
}

func GenerateUser() {
}
