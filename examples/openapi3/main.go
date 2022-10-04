package main

import (
	"log"
	"net/http"

	"github.com/FireTail-io/firetail-go-lib"
)

func health(w http.ResponseWriter, r *http.Request) {
	res := "I'm Healthy!"
	log.Println(res)
	w.WriteHeader(200)
	w.Header().Add("example-header", "example-value")
	w.Write([]byte(res))
}

func main() {
	healthHandler := http.HandlerFunc(health)
	http.Handle("/", firetail.Middleware(healthHandler))
	http.ListenAndServe(":8080", nil)
}
