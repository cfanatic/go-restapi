package restapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cfanatic/go-netchat/internal/database"
	Log "github.com/cfanatic/go-netchat/internal/logger"
	"github.com/cfanatic/go-netchat/internal/settings"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Message struct {
	Header string `json:"header"`
	Body   string `json:"body"`
}

type Request struct {
	Method  string
	Url     string
	Message Message
}

var (
	config     settings.Token
	db         *database.Database
	secret_key = []byte(config.GetSecretKey())
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func init() {
	var err error
	if db, err = database.New(); err != nil {
		Log.Log.Println(err)
		panic(err)
	}
}

func LogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Log.Log.Println(fmt.Sprintf("Request from %s to %s %s",
				strings.Split(r.RemoteAddr, ":")[0],
				r.Method,
				r.RequestURI,
			))
			next.ServeHTTP(w, r)
		},
	)
}

func AuthenticationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie("token"); err != nil {
				if err == http.ErrNoCookie {
					var cred *database.Credential
					params := mux.Vars(r)
					user, ok_user := params["user"]
					password, ok_password := params["password"]
					if cred, err = db.GetUser(user); err != nil {
						Log.Log.Println(err)
						http.Error(w, err.Error(), http.StatusUnauthorized)
						return
					}
					hash := cred.Password
					err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
					if !ok_user || !ok_password || len(hash) == 0 || err == bcrypt.ErrMismatchedHashAndPassword {
						Log.Log.Println(fmt.Sprintf("Authentification for %s failed", strings.Split(r.RemoteAddr, ":")[0]))
						http.Error(w, "Authentification failed", http.StatusUnauthorized)
						return
					}
					expiration := time.Now().Add(config.GetExpiration() * time.Minute)
					claims := &Claims{
						Username: user,
						StandardClaims: jwt.StandardClaims{
							ExpiresAt: jwt.At(expiration),
						},
					}
					tmp := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
					token, err := tmp.SignedString(secret_key)
					if err != nil {
						Log.Log.Println("Could not create token")
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
					Log.Log.Println(fmt.Sprintf(`User "%s" logged in`, user))
					next.ServeHTTP(w, r)
				} else {
					Log.Log.Println("Bad request")
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
						Log.Log.Println("Token signature invalid")
						http.Error(w, "Token signature invalid", http.StatusUnauthorized)
					} else {
						Log.Log.Println("Bad request")
						http.Error(w, "Bad request", http.StatusBadRequest)
					}
					return
				}
				if !token.Valid {
					Log.Log.Println("Authentification failed")
					http.Error(w, "Authentification failed", http.StatusUnauthorized)
					return
				}
				Log.Log.Println(fmt.Sprintf("%s authorized by existing cookie", claims.Username))
				next.ServeHTTP(w, r)
			}
		},
	)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	message, _ := unmarshall(body)
	body, _ = json.Marshal(message)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged in successfully"))
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	claims, _ := claim(r)
	body, _ := json.Marshal(Message{
		Header: "user",
		Body:   claims.Username,
	})
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func UnavailableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func SendRequest(request Request) {
	var body []byte
	time.AfterFunc(2*time.Second, func() {
		body, _ = marshall(request.Message)
		req, err := http.NewRequest(request.Method, request.Url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Do(req)
		if err != nil {
			Log.Log.Println(err)
		}
		defer resp.Body.Close()

		Log.Log.Println("Status: " + resp.Status)
		for key, value := range resp.Header {
			Log.Log.Println(fmt.Sprintf("%s: %s", key, value[0]))
		}
		body, _ = ioutil.ReadAll(resp.Body)
		if message, err := unmarshall(body); err == nil {
			Log.Log.Println(message)
		}
	})
}

func marshall(message Message) ([]byte, error) {
	var body []byte
	var err error
	if body, err = json.Marshal(message); err == nil {
		return body, nil
	} else {
		return body, err
	}
}

func unmarshall(body []byte) (Message, error) {
	var message Message
	var err error
	if err = json.Unmarshal(body, &message); err == nil {
		return message, nil
	} else {
		return message, err
	}
}

func claim(r *http.Request) (Claims, error) {
	var claims Claims
	if c, err := r.Cookie("token"); err == nil {
		token := c.Value
		keyFunc := func(token *jwt.Token) (interface{}, error) {
			return secret_key, nil
		}
		if _, err := jwt.ParseWithClaims(token, &claims, keyFunc); err == nil {
			return claims, nil
		} else {
			return claims, err
		}
	} else {
		return claims, err
	}
}
