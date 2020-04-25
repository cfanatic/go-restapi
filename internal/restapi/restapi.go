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

func Marshall(message Message) ([]byte, error) {
	var body []byte
	var err error
	if body, err = json.Marshal(message); err != nil {
		// do nothing
	}
	return body, err
}

func Unmarshall(body []byte) (Message, error) {
	var message Message
	var err error
	if err = json.Unmarshal(body, &message); err != nil {
		// do nothing
	}
	return message, err
}

func SendRequest(request Request) {
	var body []byte
	time.AfterFunc(2*time.Second, func() {
		body, _ = Marshall(request.Message)
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
		if message, err := Unmarshall(body); err == nil {
			log.Println(message)
		}
	})
}
