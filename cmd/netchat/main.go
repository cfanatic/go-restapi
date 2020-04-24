package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/cfanatic/go-netchat/internal/restapi"
	"github.com/cfanatic/go-netchat/internal/settings"
	"github.com/gorilla/mux"
)

func record(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println(fmt.Sprintf("Request as %s from %s to %s %s",
				r.Header["X-Session-Token"],
				r.RemoteAddr,
				r.Method,
				r.RequestURI,
			))
			next.ServeHTTP(w, r)
		},
	)
}

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			user := make(map[string]string)
			user["fake_token"] = "RandomUser"
			token := r.Header.Get("X-Session-Token")
			if user, found := user[token]; found {
				next.ServeHTTP(w, r)
			} else {
				log.Println("Authentification failed", user)
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		},
	)
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "GET called"})
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "POST called"})
}

func unavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"message": ""})
}

func login(w http.ResponseWriter, r *http.Request) {
	var user, password string
	var ok bool
	pathParams := mux.Vars(r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	if user, ok = pathParams["user"]; ok {
		if len(user) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "user missing"})
			return
		}
	}
	if password, ok = pathParams["password"]; ok {
		if len(password) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "password missing"})
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("{user: %s, password: %s}", user, password),
	})
}

func main() {
	addr := settings.Address()
	port := settings.PortTLS()
	router := mux.NewRouter()
	s := router.Host(addr).Subrouter()
	s.HandleFunc("/", get).Methods("GET")
	s.HandleFunc("/", post).Methods("POST")
	s.HandleFunc("/", unavailable)
	s.HandleFunc("/login/{user}:{password}", login).Methods("GET")
	s.Use(record, authenticate)
	srv := &http.Server{
		Handler:      s,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	restapi.SendRequest(restapi.Request{
		Token:  "fake_token",
		Method: "GET",
		Url:    "https://127.0.0.1/login/RandomUser:HASH",
		Body:   "Send POST request",
	})

	path_crt, _ := filepath.Abs("misc/server.crt")
	path_key, _ := filepath.Abs("misc/server.key")
	log.Fatal(srv.ListenAndServeTLS(path_crt, path_key))
}
