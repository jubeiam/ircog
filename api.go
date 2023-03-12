package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func pingController(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")

	io.WriteString(w, `{"ping": "pong"}`)
}

func helloController(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")
	r.ParseForm()

	hello := r.Form.Get("hello")

	io.WriteString(w, fmt.Sprintf(`{"hello": "%s"}`, hello))
}

func apiRouter(router *mux.Router) {
	router.HandleFunc("/ping", pingController)
	router.HandleFunc("/hello", helloController)
}
