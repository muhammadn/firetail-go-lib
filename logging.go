package firetail

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// WIP
// type Logger struct {
// 	queue        chan *LogEntry // A channel down which LogEntrys will be queued to be sent
// 	queueSize    *int64         // The number of log entries in the queue
// 	endpoint     string         // The endpoint to which logs will be sent in batches
// 	maxBatchSize int            // The maximum size of a batch
// 	maxLogAge    time.Duration  // The maximum age of a log item to hold onto
// }

// func (l *Logger) enqueue(logEntry *LogEntry) {
// 	l.queue <- logEntry
// 	atomic.AddInt64(l.queueSize, 1)
// }

// TODO: refactor
func logRequest(r *http.Request, w *draftResponseWriter, requestBody []byte, executionTime time.Duration, options *MiddlewareOptions) {
	// Create our payload to send to the firetail logging endpoint
	logPayload := LogEntry{
		Version:       The100Alpha,
		DateCreated:   time.Now().UnixMilli(),
		ExecutionTime: float64(executionTime.Milliseconds()),
		Request: Request{
			HTTPProtocol: HTTPProtocol(r.Proto),
			URI:          "", // We'll fill this in later.
			Headers:      r.Header,
			Method:       Method(r.Method),
			Body:         string(requestBody),
			IP:           "", // We'll fill this in later.
		},
		Response: Response{
			StatusCode: int64(w.statusCode),
			Body:       string(w.responseBody),
			Headers:    w.Header(),
		},
	}
	if r.TLS != nil {
		logPayload.Request.URI = "https://" + r.Host + r.URL.RequestURI()
	} else {
		logPayload.Request.URI = "http://" + r.Host + r.URL.RequestURI()
	}
	if options.SourceIPCallback != nil {
		logPayload.Request.IP = options.SourceIPCallback(r)
	} else {
		logPayload.Request.IP = strings.Split(r.RemoteAddr, ":")[0]
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
}
