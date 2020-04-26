package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	. "github.com/cfanatic/go-netchat/internal/restapi"
	"github.com/cfanatic/go-netchat/internal/settings"
	"github.com/gorilla/mux"
)

func main() {
	// load configuration parameters
	addr := settings.Address()
	port := settings.PortTLS()
	cert_crt, cert_key := settings.Certificate()
	path_crt, _ := filepath.Abs(cert_crt)
	path_key, _ := filepath.Abs(cert_key)

	// match route requests to handlers
	router := mux.NewRouter()
	s := router.Host(addr).Subrouter()
	s.HandleFunc("/", RootHandler).Methods("GET")
	s.HandleFunc("/login/{user}/{password}", LoginHandler).Methods("GET")
	s.HandleFunc("/user/me", UserHandler).Methods("GET")
	s.Use(LogHandler, AuthenticationHandler)

	// send test request after a delay of two seconds
	SendRequest(Request{
		Method: "GET",
		Url:    "https://127.0.0.1/login/RandomUser1/test1",
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
