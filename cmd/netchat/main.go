package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cfanatic/go-netchat/internal/database"
	Log "github.com/cfanatic/go-netchat/internal/logger"
	"github.com/cfanatic/go-netchat/internal/mode"
	. "github.com/cfanatic/go-netchat/internal/restapi"
	"github.com/cfanatic/go-netchat/internal/settings"
	"github.com/gorilla/mux"
)

var configB settings.Backend

func main() {
	switch mode := mode.GetMode(); mode {
	case "init":
		var user, hostname, password string
		if db, err := database.New(); err != nil {
			Log.Log.Println(err)
			panic(err)
		} else {
			fmt.Print("Enter user: ")
			fmt.Scanln(&user)
			fmt.Print("Enter hostname: ")
			fmt.Scanln(&hostname)
			fmt.Print("Enter password: ")
			fmt.Scanln(&password)
			if err = db.CreateUser(user, hostname); err != nil {
				fmt.Println(err)
			}
			if err = db.UpdateUser(user, hostname, password); err == nil {
				fmt.Println("Login Hash:", database.GenerateHash(password))
			} else {
				fmt.Println(err)
			}
		}
	case "terminal", "debug":
		// print welcome message
		Log.Log.Println("##### Starting new session #####")
		// load configuration parameters
		addr := configB.GetAddress()
		port := configB.GetPortTLS()
		cert := configB.GetCertificate()
		path_crt, path_key := cert[0], cert[1]
		// match route requests to handlers
		router := mux.NewRouter()
		s := router.Host(addr).Subrouter()
		s.HandleFunc("/", RootHandler).Methods("GET")
		s.HandleFunc("/login/{user}/{password}", LoginHandler).Methods("GET")
		s.HandleFunc("/user", UserHandler).Methods("GET")
		s.HandleFunc("/messages/{start}/{offset}", GetMessagesHandler).Methods("GET")
		s.HandleFunc("/messages/unread", GetMessagesUnreadHandler).Methods("GET")
		s.HandleFunc("/message/send", SendMessageHandler).Methods("POST")
		s.Use(LogHandler, AuthenticationHandler)
		// listen for incoming HTTPS connections
		srv := &http.Server{
			Handler:      s,
			Addr:         fmt.Sprintf(":%d", port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		log.Fatal(srv.ListenAndServeTLS(path_crt, path_key))
	}
}
