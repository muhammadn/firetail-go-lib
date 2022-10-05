package main

import (
	"net/http"
	"time"

	"github.com/FireTail-io/firetail-go-lib"
)

func health(w http.ResponseWriter, r *http.Request) {
	// Firetail will log the execution time, let's pretend this endpoint takes about 50ms...
	time.Sleep(50 * time.Millisecond)

	// Firetail will also log the response code...
	w.WriteHeader(200)

	// Firetail will also capture response headers...
	w.Header().Add("Content-Type", "text/html")

	// And, finally, it'll also log the response body...
	w.Write([]byte("I'm Healthy, and I take some time!"))
}

func main() {
	// We can setup our handler as usual, just wrap the firetail middleware around it :)
	healthHandler := firetail.Middleware(http.HandlerFunc(health))
	http.Handle("/", healthHandler)
	http.ListenAndServe(":8080", nil)
}
