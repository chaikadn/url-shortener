package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

var table = make(map[string]string)

func main() {

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	http.HandleFunc("/", postHandler)
	http.HandleFunc("/{id}", getHandler)

	fmt.Println("service started")

	return http.ListenAndServe(":8080", nil)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := generateURL()
	//TODO: handle error
	longURL, _ := io.ReadAll(r.Body)

	// FIXME: fix a possible overwrite
	table[shortURL] = string(longURL)

	w.WriteHeader(http.StatusCreated)

	//TODO: handle error
	_, _ = w.Write([]byte("http://localhost:8080/" + shortURL)) // http://localhost:8080/EwHXdJfB

	fmt.Println(table)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	location, ok := table[r.URL.Path[1:]]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func generateURL() string {
	chars := []rune("ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxuz1234567890")
	res := make([]rune, 8)

	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return string(res)
}
