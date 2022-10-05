package firetail

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

type MiddlewareOptions struct {
	SpecPath          string                       // Path at which an openapi spec can be found
	SourceIPCallback  func(r *http.Request) string // An optional callback which takes the http.Request and returns the source IP of the request as a string.
	DisableValidation *bool                        // An optional flag to disable request & response validation; validation is enabled by default
}

// GetFiretailMiddleware creates & returns a firetail middleware which will use the openapi3 spec found at `appSpecPath`.
func GetMiddleware(options *MiddlewareOptions) (func(next http.Handler) http.Handler, error) {
	// Load in our appspec, validate it & create a router from it.
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(options.SpecPath)
	if err != nil {
		return nil, err
	}
	err = doc.Validate(context.Background())
	if err != nil {
		return nil, err
	}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, err
	}

	middleware := func(next http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read in the request body so we can log it later & replace r.Body with a new copy for the next http.Handler to read from
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("Error reading in request body, err:", err.Error())
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

			// Create custom responseWriter so we can extract the response body written further down the chain
			responseWriter := &draftResponseWriter{w, 0, nil}

			// If validation has been disabled, everything is far simpler...
			if options.DisableValidation != nil && *options.DisableValidation {
				executionTime := handleRequestWithoutValidation(responseWriter, r, next)
				responseWriter.Publish()
				logRequest(r, responseWriter, requestBody, executionTime, options)
			}

			// If the request validation hasn't been disabled, then we handle the request with validation...
			executionTime, err := handleRequestWithValidation(responseWriter, r, next, router)

			// Depending upon the err we get, we may need to override the response with a particular code & body
			switch err {
			case routers.ErrPathNotFound:
				w.WriteHeader(404)
				w.Write([]byte("404 - Not Found"))
				return
			case RequestValidationError:
				w.WriteHeader(400)
				w.Write([]byte("400 - Bad Request"))
				return
			case ResponseValidationError:
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			case nil:
				break
			default:
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			}

			// If the response passed the validation, we can now publish it
			responseWriter.Publish()

			// And, finally, log it :)
			logRequest(r, responseWriter, requestBody, executionTime, options)
		})

		return handler
	}

	return middleware, nil
}

func handleRequestWithoutValidation(w *draftResponseWriter, r *http.Request, next http.Handler) time.Duration {
	startTime := time.Now()
	next.ServeHTTP(w, r)
	return time.Since(startTime)
}

func handleRequestWithValidation(w *draftResponseWriter, r *http.Request, next http.Handler, router routers.Router) (time.Duration, error) {
	// Check there's a corresponding route for this request
	route, pathParams, err := router.FindRoute(r)
	if err == routers.ErrPathNotFound {
		return time.Duration(0), err
	} else if err != nil {
		return time.Duration(0), err
	}

	// Validate the request against the OpenAPI spec.
	// We'll also need the requestValidationInput again later when validating the response.
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}
	err = openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
	if err != nil {
		return time.Duration(0), RequestValidationError
	}

	// Serve the next handler down the chain & take note of the execution time
	startTime := time.Now()
	next.ServeHTTP(w, r)
	executionTime := time.Since(startTime)

	// Validate the response against the openapi spec
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
		},
		Status: w.statusCode,
		Header: w.Header(),
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
		},
	}
	responseValidationInput.SetBodyBytes(w.responseBody)
	err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
	if err != nil {
		return time.Duration(0), ResponseValidationError
	}

	return executionTime, nil
}
