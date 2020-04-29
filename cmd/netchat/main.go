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

var configGeneral settings.General
var configBackend settings.Backend

func main() {
	// print welcome message
	Log.Log.Println("##### Starting new session #####")

	// load configuration parameters
	cred := configGeneral.GetTestUser()
	addr := configBackend.GetAddress()
	port := configBackend.GetPortTLS()
	cert := configBackend.GetCertificate()
	path_crt, path_key := cert[0], cert[1]

	// match route requests to handlers
	router := mux.NewRouter()
	s := router.Host(addr).Subrouter()
	s.HandleFunc("/", RootHandler).Methods("GET")
	s.HandleFunc("/login/{user}/{password}", LoginHandler).Methods("GET")
	s.HandleFunc("/user", UserHandler).Methods("GET")
	s.Use(LogHandler, AuthenticationHandler)

	// send test request after a delay of two seconds
	SendRequest(Request{
		Method: "GET",
		Url:    fmt.Sprintf("https://127.0.0.1:1025/login/%s/%s", cred[0], cred[1]),
		Message: Message{
			Header: "message",
			Body:   "test object",
		},
	})

	// listen for incoming HTTPS connections
	srv := &http.Server{
		Handler:      s,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServeTLS(path_crt, path_key))
}
