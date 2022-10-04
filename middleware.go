package firetail

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read in the request body & replace r.Body with a new copy for the next http.Handler to read from
		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading in request body, err:", err.Error())
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		// TODO: validate the request against the openapi spec

		// Create custom responseWriter so we can extract the response body
		responseWriter := &customResponseWriter{w, 0, nil}

		// Serve the next handler down the chain & take note of the execution time
		startTime := time.Now()
		next.ServeHTTP(responseWriter, r)
		executionTime := time.Since(startTime)

		// TODO: validate the response against the openapi spec

		// Create our payload to send to the firetail logging endpoint
		logPayload := LoggingPayload{
			Version:        "",
			DateCreated:    time.Now().UnixMilli(),
			Execution_time: executionTime.Milliseconds(),
			Source_code:    "",
			Req: RequestPayload{
				HttpProtocol: r.Proto,
				Url:          r.Host + r.URL.Path + "?" + r.URL.RawQuery,
				Headers:      r.Header,
				Method:       r.Method,
				Body:         string(requestBody),
				Ip:           r.RemoteAddr, // TODO: what if the req is proxied? Should allow custom func? E.g. to use X-Forwarded-For header etc.
			},
			Resp: ResponsePayload{
				Status_code: responseWriter.statusCode,
				Content_len: len(responseWriter.responseBody),
				Body:        string(responseWriter.responseBody),
				Headers:     responseWriter.Header(),
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
}
