package firetail

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Read in the body from r.Body using a teereader
		requestBody := ""

		// TODO: validate the request against the openapi spec

		// TODO: replace w (http.ResponseWriter) with custom responsewriter struct which keeps note of the status code

		// Serve the next handler down the chain & take note of the execution time
		startTime := time.Now()
		next.ServeHTTP(w, r)
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
				Url:          "", // TODO
				Headers:      r.Header,
				Path:         r.URL.Path,
				Method:       r.Method,
				OPath:        "",            // TODO
				FPath:        "",            // TODO
				Args:         r.URL.Query(), // TODO: check matches python lib
				Body:         requestBody,
				Ip:           r.RemoteAddr, // TODO: what if the req is proxied? Should allow custom func? E.g. to use X-Forwarded-For header etc.
				PathParams:   "",
			},
			Resp: ResponsePayload{
				Status_code:     0,
				Content_len:     0,
				Content_enc:     "",
				Failed_res_body: "",
				Body:            "",
				Headers:         map[string][]string{},
				Content_type:    "",
			},
		}

		// Marshall the payload to bytes
		reqBytes, err := json.Marshal(logPayload)
		if err != nil {
			log.Println("Err marshalling requestPayload to bytes, err:", err.Error())
		}

		// TODO: queue to be sent to logging endpoint.
		log.Println("Request body to be sent to Firetail logging endpoint:", string(reqBytes))
	})
}
