package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cfanatic/go-netchat/internal/settings"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	db    *sql.DB
	table string
	cred  *[]Credential
}

type Message struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	Message string    `json:"msg"`
	ReadL   int       `json:"read_local"`
	ReadR   int       `json:"read_remote"`
}

type Credential struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Password string `json:"password"`
}

var (
	configM settings.Mysql
	configT settings.Token
)

func New() (*Database, error) {
	var (
		db  Database
		err error
	)
	dataSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
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
	db.table = configM.GetTable()
	return &db, err
}

func (db *Database) GetUser(name string) (*Credential, error) {
	var err error
	query := &(sql.Rows{})
	cred := Credential{}
	if query, err = db.db.Query("SELECT * FROM users WHERE host=? OR name=?", name, name); err == nil {
		for query.Next() {
			err = query.Scan(&cred.ID, &cred.Name, &cred.Host, &cred.Password)
		}
		if cred == (Credential{}) {
			err = errors.New("User is not available in database")
			return &cred, err
		}
	}
	return &cred, err
}

func (db *Database) GetMessages(start, offset int, name string) (*[]Message, error) {
	var err error
	query := &(sql.Rows{})
	list := []Message{}
	if query, err = db.db.Query(
		fmt.Sprintf("SELECT * FROM %s ORDER BY date DESC LIMIT ?, ?", db.table),
		start,
		offset,
	); err == nil {
		message := Message{}
		for query.Next() {
			err = query.Scan(
				&message.ID,
				&message.Name,
				&message.Date,
				&message.Message,
				&message.ReadL,
				&message.ReadR,
			)
			list = append(list, message)
			if err = db.UpdateMessage(name, message); err != nil {
				return &list, err
			}
		}
	}
	return &list, err
}

func (db *Database) GetMessagesUnread(name string) (*[]Message, error) {
	var err error
	query := &(sql.Rows{})
	list := []Message{}
	if query, err = db.db.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE (read_local=0 AND name=?) OR (read_remote=0 AND name!=?)", db.table),
		name,
		name,
	); err == nil {
		message := Message{}
		for query.Next() {
			err = query.Scan(
				&message.ID,
				&message.Name,
				&message.Date,
				&message.Message,
				&message.ReadL,
				&message.ReadR,
			)
			list = append(list, message)
			if err = db.UpdateMessage(name, message); err != nil {
				return &list, err
			}
		}
	}
	return &list, err
}

func (db *Database) UpdateMessage(name string, message Message) error {
	var err error
	if message.Name == name {
		if _, err = db.db.Exec(
			fmt.Sprintf("UPDATE %s SET read_local=1 WHERE id=?", db.table), message.ID); err == nil {
		}
	} else {
		if _, err = db.db.Exec(
			fmt.Sprintf("UPDATE %s SET read_remote=1 WHERE id=?", db.table), message.ID); err == nil {
		}
	}
	return err
}

func (db *Database) GetMessageCount() (uint, error) {
	var (
		count uint
		err   error
	)
	query := &(sql.Rows{})
	if query, err = db.db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s", db.table)); err == nil {
		for query.Next() {
			err = query.Scan(&count)
		}
	}
	return count, err
}

func (db *Database) SendMessage(message Message) error {
	var (
		res sql.Result
		err error
	)
	if res, err = db.db.Exec(
		fmt.Sprintf("INSERT INTO %s (name, date, msg, read_local, read_remote) VALUES (?, ?, ?, ?, ?)", db.table),
		message.Name,
		message.Date,
		message.Message,
		message.ReadL,
		message.ReadR,
	); err == nil {
		if cnt, err := res.RowsAffected(); err == nil {
			if cnt != 1 {
				err = errors.New("Could not insert message")
				return err
			}
		}
	}
	return err
}

func (db *Database) CreateUser(name, hostname string) error {
	var (
		res sql.Result
		err error
	)
	if _, err = db.GetUser(name); err == nil {
		err = errors.New("User already available")
		return err
	}
	if res, err = db.db.Exec(
		fmt.Sprintf("INSERT INTO %s (name, host, password) VALUES (?, ?, ?)", "users"),
		name,
		hostname,
		"",
	); err == nil {
		if cnt, err := res.RowsAffected(); err == nil {
			if cnt != 1 {
				err = errors.New("Could not create new user")
				return err
			}
		}
	}
	return err
}

func (db *Database) UpdateUser(name, hostname, password string) error {
	var (
		hash  []byte
		query *sql.Rows
		res   sql.Result
		err   error
	)
	tmp := []byte(GenerateHash(password))
	if hash, err = bcrypt.GenerateFromPassword(tmp, bcrypt.DefaultCost); err == nil {
		if query, err = db.db.Query("SELECT EXISTS(SELECT 1 FROM users WHERE name=?)", name); err == nil {
			var cnt int
			query.Next()
			query.Scan(&cnt)
			if cnt == 0 {
				err = errors.New("User is not available in database: " + name)
				return err
			}
		}
		if res, err = db.db.Exec("UPDATE users SET host=?, password=? WHERE name=?", hostname, hash, name); err == nil {
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
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", password, salt)))
	return hex.EncodeToString(hash.Sum(nil))
}
