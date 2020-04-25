package restapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

func Marshall(message Message) []byte {
	var body []byte
	var err error
	if body, err = json.Marshal(message); err != nil {
		panic("Could not marshal JSON string")
	}
	return body
}

func Unmarshall(body []byte) Message {
	var message Message
	if len(body) > 0 {
		if err := json.Unmarshal(body, &message); err != nil {
			panic("Could not unmarshal JSON string")
		}
	}
	return message
}

func SendRequest(request Request) {
	var body []byte
	time.AfterFunc(2*time.Second, func() {
		body = Marshall(request.Message)
		req, err := http.NewRequest(request.Method, request.Url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		log.Println("Status:", resp.Status)
		for key, value := range resp.Header {
			log.Println(fmt.Sprintf("%s: %s", key, value[0]))
		}
		body, _ = ioutil.ReadAll(resp.Body)
		message := Unmarshall(body)
		log.Println(message)
	})
}
