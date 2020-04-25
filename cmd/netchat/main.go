package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var database = map[string]string{
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
			user, ok_user := params["user"]
			password, ok_password := params["password"]
			if !ok_user || !ok_password || password != database[user] {
				log.Println("Authentification failed")
				http.Error(w, "Authentification failed", http.StatusUnauthorized)
				return
			}
			expiration := time.Now().Add(settings.Expiration() * time.Minute)
			claims := &Claims{
				Username: user,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: jwt.At(expiration),
				},
			}
			tmp := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			token, err := tmp.SignedString(secret_key)
			if err != nil {
				log.Println("Could not create JWT")
				http.Error(w, "Could not create JWT", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "token",
				Value:   token,
				Expires: expiration,
			})
			log.Println(fmt.Sprintf("%s authorized", user))
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
	var message restapi.Message
	body, _ := ioutil.ReadAll(r.Body)
	if len(body) > 0 {
		if err := json.Unmarshal(body, &message); err != nil {
			log.Println("Could not unmarshal JSON string")
		} else {
			log.Println("Body: ", message)
		}
	}
	message.Header, message.Body = "message", "this is a test"
	buf, _ := json.Marshal(message)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
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

	message := restapi.Message{
		Header: "message",
		Body:   "test object",
	}
	restapi.SendRequest(restapi.Request{
		Method:  "GET",
		Url:     "https://127.0.0.1/login/RandomUser1/test1",
		Message: message,
	})

	path_crt, _ := filepath.Abs(cert_crt)
	path_key, _ := filepath.Abs(cert_key)
	log.Fatal(srv.ListenAndServeTLS(path_crt, path_key))
}
