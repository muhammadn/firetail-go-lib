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

			// Create a Firetail ResponseWriter so we can access the response body, status code etc. for logging & validation later
			localResponseWriter := &utils.ResponseWriter{ResponseWriter: w, StatusCode: 0, ResponseBody: nil}

			// No matter what happens, read the response from the response writer, enqueue the log entry & publish the response that was written to the ResponseWriter
			defer func() {
				logEntry.Response = logging.Response{
					StatusCode: int64(localResponseWriter.StatusCode),
					Body:       string(localResponseWriter.ResponseBody),
					Headers:    localResponseWriter.Header(),
				}
				batchLogger.Enqueue(&logEntry)
				localResponseWriter.Publish()
			}()

			// Read in the request body so we can log it & replace r.Body with a new copy for the next http.Handler to read from
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				options.ErrHandler(err, localResponseWriter)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

			// Now we have the request body, we can fill it into our log entry
			logEntry.Request.Body = string(requestBody)

			// Check there's a corresponding route for this request
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				options.ErrHandler(err, localResponseWriter)
				return
			}

			// We now know the resource that was requested, so we can fill it into our log entry
			logEntry.Request.Resource = route.Path

			// If validation has been disabled, everything is far simpler...
			if options.DisableValidation != nil && *options.DisableValidation {
				executionTime := handleWithoutValidation(localResponseWriter, r, next)
				logEntry.ExecutionTime = float64(executionTime)
				return
			}

			// If the request validation hasn't been disabled, then we handle the request with validation
			chainResponseWriter := &utils.ResponseWriter{ResponseWriter: localResponseWriter, StatusCode: 0, ResponseBody: nil}
			executionTime, err := handleWithValidation(chainResponseWriter, r, next, route, pathParams)

			// We now know the execution time, so we can fill it into our log entry
			logEntry.ExecutionTime = float64(executionTime)

			// Depending upon the err we get, we may need to override the response with a particular code & body
			if err != nil {
				options.ErrHandler(err, localResponseWriter)
				return
			}

			// If there's no err, then we can publish the response written down the chain to our localResponseWriter
			chainResponseWriter.Publish()
		})
	}

	return middleware, nil
}
