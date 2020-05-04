package restapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
	Name string `json:"name"`
	Date string `json:"date"`
	Text string `json:"text"`
}

type Messages []Message

type Request struct {
	Method  string
	Url     string
	Message Message
}

var (
	db        *database.Database
	configT   settings.Token
	secretKey = []byte(configT.GetSecretKey())
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
					if !ok_user || !ok_password {
						Log.Log.Println(fmt.Sprintf("Error: Login information missing for %s", strings.Split(r.RemoteAddr, ":")[0]))
						http.Error(w, `{"error":"login information missing"}`, http.StatusUnauthorized)
						return
					}
					if cred, err = db.GetUser(user); err != nil {
						Log.Log.Println("Error:", err, fmt.Sprintf("for %s", strings.Split(r.RemoteAddr, ":")[0]))
						http.Error(w, `{"error":"unknown user"}`, http.StatusUnauthorized)
						return
					}
					hash := cred.Password
					err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
					if len(hash) == 0 || err == bcrypt.ErrMismatchedHashAndPassword {
						Log.Log.Println(fmt.Sprintf("Error: Authentification failed for %s", strings.Split(r.RemoteAddr, ":")[0]))
						http.Error(w, `{"error":"authentification failed"}`, http.StatusUnauthorized)
						return
					}
					expiration := time.Now().Add(configT.GetExpiration() * time.Minute)
					claims := &Claims{
						Username: user,
						StandardClaims: jwt.StandardClaims{
							ExpiresAt: jwt.At(expiration),
						},
					}
					tmp := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
					token, err := tmp.SignedString(secretKey)
					if err != nil {
						Log.Log.Println("Error: Could not create token")
						http.Error(w, `{"error":"could not create token"}`, http.StatusInternalServerError)
						return
					}
					http.SetCookie(w, &http.Cookie{
						Name:     "token",
						Value:    token,
						Expires:  expiration,
						HttpOnly: true,
						Path:     "/",
					})
					next.ServeHTTP(w, r)
				} else {
					Log.Log.Println("Error: Bad request")
					http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
					return
				}
			} else {
				claims := &Claims{}
				token, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
					return secretKey, nil
				})
				if err != nil {
					if err == jwt.ErrSignatureInvalid {
						Log.Log.Println("Error: Token signature invalid")
						http.Error(w, `{"error":"token signature invalid"}`, http.StatusUnauthorized)
					} else {
						Log.Log.Println("Error: Bad request")
						http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
					}
					return
				}
				if !token.Valid {
					Log.Log.Println("Error: Authentification failed")
					http.Error(w, `{"error":"authentification failed"}`, http.StatusUnauthorized)
					return
				}
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
	body, _ := json.Marshal(map[string]string{"status": "login successful"})
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	claims, _ := claim(r)
	body, _ := json.Marshal(map[string]string{"user": claims.Username})
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	params := mux.Vars(r)
	start, ok_start := params["start"]
	offset, ok_offset := params["offset"]
	if !ok_start || !ok_offset {
		body, _ = json.Marshal(map[string]string{"error": "parameters missing to get messages"})
		w.WriteHeader(http.StatusBadRequest)
	} else {
		startInt, _ := strconv.Atoi(start)
		offsetInt, _ := strconv.Atoi(offset)
		if res, err := db.GetMessages(startInt, offsetInt); err != nil {
			body, _ = json.Marshal(map[string]string{"error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			var list Messages
			for _, item := range *res {
				message := Message{
					Name: item.Name,
					Date: string(item.Date),
					Text: item.Message,
				}
				list = append(list, message)
			}
			body, _ = json.Marshal(list)
			w.WriteHeader(http.StatusOK)
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func GetMessagesUnreadHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	claims, _ := claim(r)
	if res, err := db.GetMessagesUnread(claims.Username); err != nil {
		body, _ = json.Marshal(map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		var list Messages
		for _, item := range *res {
			message := Message{
				Name: item.Name,
				Date: string(item.Date),
				Text: item.Message,
			}
			list = append(list, message)
		}
		body, _ = json.Marshal(list)
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func UnavailableHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := json.Marshal(map[string]string{"status": "handler not allowed"})
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(body)
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
			return secretKey, nil
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
