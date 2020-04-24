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

type Request struct {
	Token  string
	Method string
	Url    string
	Body   string
}

func SendRequest(request Request) string {
	var message string
	time.AfterFunc(2*time.Second, func() {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(map[string]string{"message": request.Body})
		req, err := http.NewRequest(request.Method, request.Url, buf)
		req.Header.Set("X-Session-Token", request.Token)
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
		var body map[string]interface{}
		tmp, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(tmp, &body); err != nil {
			log.Println("Could not unmarshal JSON string")
		} else {
			if message = body["message"].(string); len(message) > 0 {
				log.Print("Body: ", message)
			}

		}
	})
	return message
}
