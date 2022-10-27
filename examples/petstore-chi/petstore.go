// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/FireTail-io/firetail-go-lib/examples/petstore-chi/api"
	firetail "github.com/FireTail-io/firetail-go-lib/middlewares/http"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
)

func main() {
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	petStore := api.NewPetStore()

	// This is how you set up a basic chi router
	r := chi.NewRouter()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
		OpenapiSpecPath: "./petstore-expanded.yaml",
		AuthCallbacks: map[string]openapi3filter.AuthenticationFunc{
			"MyBearerAuth": func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
				authHeaderValue := ai.RequestValidationInput.Request.Header.Get("Authorization")
				if authHeaderValue == "" {
					return errors.New("no bearer token supplied for \"MyBearerAuth\"")
				}
				tokenParts := strings.Split(authHeaderValue, " ")
				if len(tokenParts) != 2 || strings.ToUpper(tokenParts[0]) != "BEARER" || tokenParts[1] != "header.payload.signature" {
					return errors.New("invalid bearer token for \"MyBearerAuth\"")
				}
				return nil
			},
		},
	},
	)
	if err != nil {
		panic(err)
	}

	r.Use(firetailMiddleware)

	// We now register our petStore above as the handler for the interface
	api.HandlerFromMux(petStore, r)

	s := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", *port),
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
