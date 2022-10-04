package main

import (
	"log"
	"net/http"
	"time"

	"github.com/FireTail-io/firetail-go-lib"
)

func health(w http.ResponseWriter, r *http.Request) {
	res := "I'm Healthy, and I take some time!"
	log.Println(res)
	// Firetail will log the execution time, let's pretend this endpoint takes about 50ms...
	time.Sleep(50 * time.Millisecond)
	w.WriteHeader(200)
	w.Header().Add("example-header", "example-value")
	w.Write([]byte(res))
}

func main() {
	healthHandler := http.HandlerFunc(health)
	http.Handle("/", firetail.Middleware(healthHandler))
	http.ListenAndServe(":8080", nil)
}
