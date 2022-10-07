package firetail

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FireTail-io/firetail-go-lib/logging"
	"github.com/FireTail-io/firetail-go-lib/utils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// Options is an options struct used when creating a Firetail middleware (GetMiddleware)
type Options struct {
	SpecPath          string                       // Path at which an openapi spec can be found
	SourceIPCallback  func(r *http.Request) string // An optional callback which takes the http.Request and returns the source IP of the request as a string.
	DisableValidation *bool                        // An optional flag to disable request & response validation; validation is enabled by default
	FiretailEndpoint  string                       // The Firetail logging endpoint request data should be sent to
}

// GetMiddleware creates & returns a firetail middleware. Errs if the openapi spec can't be found, validated, or loaded into a gorillamux router.
func GetMiddleware(options *Options) (func(next http.Handler) http.Handler, error) {
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

	// Create a batchLogger to pass all our log entries to
	batchLogger := logging.NewBatchLogger(1024*512, time.Second, options.FiretailEndpoint)

	middleware := func(next http.Handler) http.Handler {
		// TODO: refactor to ALWAYS send a log to Firetail - even when validation fails?
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check there's a corresponding route for this request
			route, pathParams, err := router.FindRoute(r)
			if err == routers.ErrPathNotFound {
				w.WriteHeader(404)
				w.Write([]byte("404 - Not Found"))
				return
			} else if err != nil {
				log.Println(err.Error())
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			}

			// Read in the request body so we can log it later & replace r.Body with a new copy for the next http.Handler to read from
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

			// Create a draftResponseWriter so we can extract the response body written further down the chain for validation & logging later
			draftResponseWriter := &utils.DraftResponseWriter{ResponseWriter: w, StatusCode: 0, ResponseBody: nil}

			// If validation has been disabled, everything is far simpler...
			if options.DisableValidation != nil && *options.DisableValidation {
				executionTime := handleWithoutValidation(draftResponseWriter, r, next)
				draftResponseWriter.Publish()
				logEntry := createLogEntry(r, draftResponseWriter, requestBody, route.Path, executionTime, options.SourceIPCallback)
				batchLogger.Enqueue(&logEntry)
				return
			}

			// If the request validation hasn't been disabled, then we handle the request with validation
			executionTime, err := handleWithValidation(draftResponseWriter, r, next, route, pathParams)

			// Depending upon the err we get, we may need to override the response with a particular code & body
			switch err {
			case routers.ErrPathNotFound:
				w.WriteHeader(404)
				w.Write([]byte("404 - Not Found"))
				return
			case utils.RequestValidationError:
				w.WriteHeader(400)
				w.Write([]byte("400 - Bad Request"))
				return
			case utils.ResponseValidationError:
				log.Println(err.Error())
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			case nil:
				// Happy path is nil, so just break
				break
			default:
				// If we get any other non-nil err we return a generic 500
				log.Println(err.Error())
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			}

			// If the response passed the validation, we can now publish it
			draftResponseWriter.Publish()

			// And, finally, log it :)
			logEntry := createLogEntry(r, draftResponseWriter, requestBody, route.Path, executionTime, options.SourceIPCallback)
			batchLogger.Enqueue(&logEntry)
		})

		return handler
	}

	return middleware, nil
}

func handleWithoutValidation(w *utils.DraftResponseWriter, r *http.Request, next http.Handler) time.Duration {
	// There's no validation to do; we've just got to record the execution time
	startTime := time.Now()
	next.ServeHTTP(w, r)
	return time.Since(startTime)
}

func handleWithValidation(w *utils.DraftResponseWriter, r *http.Request, next http.Handler, route *routers.Route, pathParams map[string]string) (time.Duration, error) {
	// Validate the request against the OpenAPI spec. We'll also need the requestValidationInput again later when validating the response.
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}
	err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
	if err != nil {
		return time.Duration(0), utils.RequestValidationError
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
		Status: w.StatusCode,
		Header: w.Header(),
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
		},
	}
	responseValidationInput.SetBodyBytes(w.ResponseBody)
	err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
	if err != nil {
		return time.Duration(0), utils.ResponseValidationError
	}

	return executionTime, nil
}

func createLogEntry(r *http.Request, w *utils.DraftResponseWriter, requestBody []byte, resourcePath string, executionTime time.Duration, sourceIPCallback func(r *http.Request) string) logging.LogEntry {
	// Create our payload to send to the firetail logging endpoint
	logEntry := logging.LogEntry{
		Version:       logging.The100Alpha,
		DateCreated:   time.Now().UnixMilli(),
		ExecutionTime: float64(executionTime.Milliseconds()),
		Request: logging.Request{
			HTTPProtocol: logging.HTTPProtocol(r.Proto),
			URI:          "", // We'll fill this in later.
			Headers:      r.Header,
			Method:       logging.Method(r.Method),
			Body:         string(requestBody),
			IP:           "", // We'll fill this in later.
			Resource:     resourcePath,
		},
		Response: logging.Response{
			StatusCode: int64(w.StatusCode),
			Body:       string(w.ResponseBody),
			Headers:    w.Header(),
		},
	}
	if r.TLS != nil {
		logEntry.Request.URI = "https://" + r.Host + r.URL.RequestURI()
	} else {
		logEntry.Request.URI = "http://" + r.Host + r.URL.RequestURI()
	}
	if sourceIPCallback != nil {
		logEntry.Request.IP = sourceIPCallback(r)
	} else {
		logEntry.Request.IP = strings.Split(r.RemoteAddr, ":")[0]
	}

	return logEntry
}
