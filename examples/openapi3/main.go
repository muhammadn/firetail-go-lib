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

	// Firetail will also log the response code...
	w.WriteHeader(200)

	// Firetail will also capture response headers...
	w.Header().Add("example-header", "example-value")

	// And, finally, it'll also log the response body...
	w.Write([]byte(res))
}

func main() {
	healthHandler := http.HandlerFunc(health)
	http.Handle("/", firetail.Middleware(healthHandler))
	http.ListenAndServe(":8080", nil)
}
