package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	. "github.com/cfanatic/go-netchat/internal/restapi"
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

func getClaims(token string) *Claims {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return secret_key, nil
	}
	claims := &Claims{}
	if _, err := jwt.ParseWithClaims(token, claims, keyFunc); err == nil {
		return claims
	} else {
		return nil
	}
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
			if c, err := r.Cookie("token"); err != nil {
				if err == http.ErrNoCookie {
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
						log.Println("Could not create token")
						http.Error(w, "Could not create token", http.StatusInternalServerError)
						return
					}
					http.SetCookie(w, &http.Cookie{
						Name:     "token",
						Value:    token,
						Expires:  expiration,
						HttpOnly: true,
						Path:     "/",
					})
					log.Println(fmt.Sprintf("%s logged in", user))
					next.ServeHTTP(w, r)
				} else {
					log.Println("Bad request")
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
			} else {
				claims := &Claims{}
				token, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
					return secret_key, nil
				})
				if err != nil {
					if err == jwt.ErrSignatureInvalid {
						log.Println("Token signature invalid")
						http.Error(w, "Token signature invalid", http.StatusUnauthorized)
					} else {
						log.Println("Bad request")
						http.Error(w, "Bad request", http.StatusBadRequest)
					}
					return
				}
				if !token.Valid {
					log.Println("Authentification failed")
					http.Error(w, "Authentification failed", http.StatusUnauthorized)
					return
				}
				log.Println(fmt.Sprintf("%s authorized by cookie", claims.Username))
				next.ServeHTTP(w, r)
			}
		},
	)
}

func get(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	message, _ := Unmarshall(body)
	body, _ = json.Marshal(message)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}

func unavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged in successfully"))
}

func user(w http.ResponseWriter, r *http.Request) {
	var claims *Claims
	if c, err := r.Cookie("token"); err == nil {
		claims = getClaims(c.Value)
	}
	body, _ := json.Marshal(Message{
		Header: "user",
		Body:   claims.Username,
	})
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
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
	s.HandleFunc("/user/me", user).Methods("GET")
	s.Use(record, authenticate)
	srv := &http.Server{
		Handler:      s,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	SendRequest(Request{
		Method: "GET",
		Url:    "https://127.0.0.1/login/RandomUser1/test1",
		Message: Message{
			Header: "message",
			Body:   "test object",
		},
	})

	path_crt, _ := filepath.Abs(cert_crt)
	path_key, _ := filepath.Abs(cert_key)
	log.Fatal(srv.ListenAndServeTLS(path_crt, path_key))
}
