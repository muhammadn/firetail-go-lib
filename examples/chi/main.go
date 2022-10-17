package main

import (
	"net/http"

	firetail "github.com/FireTail-io/firetail-go-lib/middlewares/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
		SpecPath: "app-spec.yaml",
	})
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(firetailMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("I'm Healthy! :)"))
	})

	r.Get("/auth-example", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("You're in! :)"))
	})

	http.ListenAndServe(":3333", r)
}
