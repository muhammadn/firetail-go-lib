package firetail

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/FireTail-io/firetail-go-lib/logging"
	"github.com/FireTail-io/firetail-go-lib/utils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// GetMiddleware creates & returns a firetail middleware. Errs if the openapi spec can't be found, validated, or loaded into a gorillamux router.
func GetMiddleware(options *Options) (func(next http.Handler) http.Handler, error) {
	// If sourceIPCallback or ErrHandler are nil, we fill them in with our own defaults
	if options.SourceIPCallback == nil {
		options.SourceIPCallback = defaultSourceIPCallback
	}
	if options.ErrHandler == nil {
		options.ErrHandler = defaultErrHandler
	}

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
	// TODO: change max log age to a minute
	batchLogger := logging.NewBatchLogger(1024*512, time.Second, options.FiretailEndpoint)

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a LogEntry populated with everything we know right now
			logEntry := logging.LogEntry{
				Version:     logging.The100Alpha,
				DateCreated: time.Now().UnixMilli(),
				Request: logging.Request{
					HTTPProtocol: logging.HTTPProtocol(r.Proto),
					Headers:      r.Header,
					Method:       logging.Method(r.Method),
					IP:           options.SourceIPCallback(r),
				},
			}
			if r.TLS != nil {
				logEntry.Request.URI = "https://" + r.Host + r.URL.RequestURI()
			} else {
				logEntry.Request.URI = "http://" + r.Host + r.URL.RequestURI()
			}

			// Create a draftResponseWriter so we can access the response body, status code etc. for logging & validation later
			localDraftResponseWriter := &utils.DraftResponseWriter{ResponseWriter: w, StatusCode: 0, ResponseBody: nil}

			// No matter what happens, read the response from the draft response writer, enqueue the log entry & publish the draft
			defer func() {
				logEntry.Response = logging.Response{
					StatusCode: int64(localDraftResponseWriter.StatusCode),
					Body:       string(localDraftResponseWriter.ResponseBody),
					Headers:    localDraftResponseWriter.Header(),
				}
				batchLogger.Enqueue(&logEntry)
				localDraftResponseWriter.Publish()
			}()

			// Read in the request body so we can log it & replace r.Body with a new copy for the next http.Handler to read from
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				options.ErrHandler(err, localDraftResponseWriter)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

			// Now we have the request body, we can fill it into our log entry
			logEntry.Request.Body = string(requestBody)

			// Check there's a corresponding route for this request
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				options.ErrHandler(err, localDraftResponseWriter)
				return
			}

			// We now know the resource that was requested, so we can fill it into our log entry
			logEntry.Request.Resource = route.Path

			// If validation has been disabled, everything is far simpler...
			if options.DisableValidation != nil && *options.DisableValidation {
				executionTime := handleWithoutValidation(localDraftResponseWriter, r, next)
				logEntry.ExecutionTime = float64(executionTime)
				return
			}

			// If the request validation hasn't been disabled, then we handle the request with validation
			chainDraftResponseWriter := &utils.DraftResponseWriter{ResponseWriter: localDraftResponseWriter, StatusCode: 0, ResponseBody: nil}
			executionTime, err := handleWithValidation(chainDraftResponseWriter, r, next, route, pathParams)

			// We now know the execution time, so we can fill it into our log entry
			logEntry.ExecutionTime = float64(executionTime)

			// Depending upon the err we get, we may need to override the response with a particular code & body
			if err != nil {
				options.ErrHandler(err, localDraftResponseWriter)
				return
			}

			// If there's no err, then we can publish the response written down the chain to our localDraftResponseWriter
			chainDraftResponseWriter.Publish()
		})
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
		return time.Duration(0), utils.ErrRequestValidationFailed
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
		return time.Duration(0), utils.ErrResponseValidationFailed
	}

	return executionTime, nil
}
