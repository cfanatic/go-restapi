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

func SendRequest(request Request) {
	time.AfterFunc(2*time.Second, func() {
		buf, err := json.Marshal(request.Message)
		req, err := http.NewRequest(request.Method, request.Url, bytes.NewBuffer(buf))
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
		var message Message
		body, _ := ioutil.ReadAll(resp.Body)
		if len(body) > 0 {
			if err := json.Unmarshal(body, &message); err != nil {
				log.Println("Could not unmarshal JSON string")
			} else {
				log.Println("Body: ", message)
			}
		}
	})
}
