package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	firetail "github.com/FireTail-io/firetail-go-lib/middlewares/http"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
		OpenapiSpecPath: "app-spec.yaml",
		AuthCallback: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
			switch ai.SecuritySchemeName {
			case "BasicAuth":
				token := ai.RequestValidationInput.Request.Header.Get("Authorization")
				// b64 encoding of 'admin:password'
				if token != "Basic YWRtaW46cGFzc3dvcmQ=" {
					return errors.New("invalid authorization token")
				}
				return nil
			case "ApiKeyAuth":
				token := ai.RequestValidationInput.Request.Header.Get("X-Api-Key")
				if token != "example-api-key" {
					return errors.New("invalid API key")
				}
				return nil
			case "BearerAuth":
				tokenParts := strings.Split(ai.RequestValidationInput.Request.Header.Get("Authorization"), " ")
				if len(tokenParts) < 2 || tokenParts[1] != "header.payload.signature" {
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
