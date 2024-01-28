package main_test

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"net/http"
	"time"
)

const baseUrl = "http://localhost:3000/object/"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var client = &http.Client{
	Timeout: 60 * time.Second,
}

func generateID() string {
	id := make([]rune, 32)

	for i := range id {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}

		id[i] = letters[num.Int64()]
	}

	return string(id)
}

func generateBody() string {
	return "body-" + generateID()
}

func httpPutObject(id string, body string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, baseUrl+id, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func httpGetObject(id string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+id, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
