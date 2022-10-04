package main

import (
	"log"
	"net/http"

	"github.com/FireTail-io/firetail-go-lib"
)

func health(w http.ResponseWriter, r *http.Request) {
	res := "I'm Healthy!"
	log.Println(res)
	w.Write([]byte(res))
	w.WriteHeader(200)
}

func main() {
	healthHandler := http.HandlerFunc(health)
	http.Handle("/", firetail.Middleware(healthHandler))
	http.ListenAndServe(":8080", nil)
}
