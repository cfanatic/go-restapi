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
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gorilla/mux"
)

var secret_key = []byte("my_secret_key")

var users = map[string]string{
	"RandomUser1": "test1",
	"RandomUser2": "test2",
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func record(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println(fmt.Sprintf("Request from %s to %s %s",
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
			params := mux.Vars(r)
			password, ok := params["password"]
			fmt.Println(params)
			if !ok || password != users[params["user"]] {
				log.Println("Authentification failed")
				http.Error(w, "Authentification failed", http.StatusUnauthorized)
			}
			expiration := time.Now().Add(5 * time.Minute)
			claims := &Claims{
				Username: params["user"],
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: jwt.At(expiration),
				},
			}
			tmp := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			token, err := tmp.SignedString(secret_key)
			if err != nil {
				log.Println("Could not create JWT")
				http.Error(w, "Could not create JWT", http.StatusInternalServerError)
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "token",
				Value:   token,
				Expires: expiration,
			})
			next.ServeHTTP(w, r)
		},
	)
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "GET called"})
}

func unavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"message": ""})
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "passed all",
	})
}

func main() {
	addr := settings.Address()
	port := settings.PortTLS()
	cert_crt, cert_key := settings.Certificate()
	router := mux.NewRouter()
	s := router.Host(addr).Subrouter()
	s.HandleFunc("/", get).Methods("GET")
	s.HandleFunc("/", unavailable)
	s.HandleFunc("/login/{user}/{password}", login).Methods("GET")
	s.Use(record, authenticate)
	srv := &http.Server{
		Handler:      s,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	restapi.SendRequest(restapi.Request{
		Method: "GET",
		Url:    "https://127.0.0.1/login/RandomUser1/test1",
		Body:   "Send POST request",
	})

	path_crt, _ := filepath.Abs(cert_crt)
	path_key, _ := filepath.Abs(cert_key)
	log.Fatal(srv.ListenAndServeTLS(path_crt, path_key))
}
