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
	w.Header().Add("Content-Type", "text/plain")

	// And, finally, it'll also log the response body...
	w.Write([]byte("I'm Healthy!"))
}

func main() {
	// Get a firetail middleware
	firetailMiddleware, err := firetail.GetFiretailMiddleware("app-spec.yaml")
	if err != nil {
		panic(err)
	}

	// We can setup our handler as usual, just wrap it in the firetailMiddleware :)
	healthHandler := firetailMiddleware(http.HandlerFunc(health))
	http.Handle("/health", healthHandler)

	http.ListenAndServe(":8080", nil)
}
