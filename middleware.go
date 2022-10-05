package firetail

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// GetFiretailMiddleware creates & returns a firetail middleware which will use the openapi3 spec found at `appSpecPath`.
func GetFiretailMiddleware(appSpecPath string) (func(next http.Handler) http.Handler, error) {
	// Load in our appspec, validate it & create a router from it.
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(appSpecPath)
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
			// Check there's a corresponding route for this request
			route, pathParams, err := router.FindRoute(r)
			if err == routers.ErrPathNotFound {
				w.WriteHeader(404)
				w.Write([]byte("404 - Not Found"))
				return
			} else if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}

			// Validate the request against the OpenAPI spec
			err = openapi3filter.ValidateRequest(
				context.Background(),
				&openapi3filter.RequestValidationInput{
					Request:    r,
					PathParams: pathParams,
					Route:      route,
				},
			)
			if err != nil {
				w.WriteHeader(400)
				w.Write([]byte("400 - Bad Request: " + err.Error()))
				return
			}

			// Read in the request body & replace r.Body with a new copy for the next http.Handler to read from
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("Error reading in request body, err:", err.Error())
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

			// Create custom responseWriter so we can extract the response body written further down the chain
			responseWriter := &customResponseWriter{w, 0, nil}

			// Serve the next handler down the chain & take note of the execution time
			startTime := time.Now()
			next.ServeHTTP(responseWriter, r)
			executionTime := time.Since(startTime)

			// TODO: validate the response against the openapi spec

			// Create our payload to send to the firetail logging endpoint
			logPayload := LogEntry{
				Version:       The100Alpha,
				DateCreated:   time.Now().UnixMilli(),
				ExecutionTime: float64(executionTime.Milliseconds()),
				Request: Request{
					HTTPProtocol: HTTPProtocol(r.Proto),
					URI:          r.Host + r.URL.Path + "?" + r.URL.RawQuery,
					Headers:      r.Header,
					Method:       Method(r.Method),
					Body:         string(requestBody),
					IP:           strings.Split(r.RemoteAddr, ":")[0], // TODO: what if the req is proxied? Should allow custom func? E.g. to use X-Forwarded-For header etc.
				},
				Response: Response{
					StatusCode: int64(responseWriter.statusCode),
					Body:       string(responseWriter.responseBody),
					Headers:    responseWriter.Header(),
				},
			}

			// Marshall the payload to bytes. Using MarshalIndent for now as we're just logging it & it makes it easier to read.
			// TODO: revert to json.Marshal when actually sending to Firetail endpoint
			reqBytes, err := json.MarshalIndent(logPayload, "", "	")
			if err != nil {
				log.Println("Err marshalling requestPayload to bytes, err:", err.Error())
				return
			}

			// TODO: queue to be sent to logging endpoint.
			log.Println("Request body to be sent to Firetail logging endpoint:", string(reqBytes))
		})

		return handler
	}

	return middleware, nil
}
