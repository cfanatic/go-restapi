package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	Log "github.com/cfanatic/go-netchat/internal/logger"
	. "github.com/cfanatic/go-netchat/internal/restapi"
	"github.com/cfanatic/go-netchat/internal/settings"
	"github.com/gorilla/mux"
)

var configB settings.Backend

func main() {
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
