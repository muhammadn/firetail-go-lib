package main

import (
	"context"
	"errors"
	"net/http"

	firetail "github.com/FireTail-io/firetail-go-lib/middlewares/http"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
		SpecPath: "app-spec.yaml",
		AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
			switch ai.SecuritySchemeName {
			case "BasicAuth":
				// TODO
				return errors.New("BasicAuth is not implemented yet")
			case "ApiKeyAuth":
				// TODO
				return errors.New("ApiKeyAuth is not implemented yet")
			case "BearerAuth":
				token := ai.RequestValidationInput.Request.Header.Get("Authorization")
				if token != "bearer example-token" {
					return errors.New("invalid bearer token")
				}
				return nil
			default:
				return errors.New("security scheme not implemented")
			}
		},
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
