package firetail

// The Payload struct holds a payload to be sent to the firetail logging endpoint
type LoggingPayload struct {
	Version       string          `json:"version"`
	DateCreated   int64           `json:"dateCreated"`   // UNIX Milliseconds
	ExecutionTime int64           `json:"executionTime"` // Milliseconds
	SourceCode    string          `json:"sourceCode"`
	Request       RequestPayload  `json:"request"`
	Response      ResponsePayload `json:"response"`
}

// The requestPayload struct holds the information about the request payload that is passed to the firetail logging endpoint
type RequestPayload struct {
	HttpProtocol string              `json:"httpProtocol"`
	Url          string              `json:"url"`
	Headers      map[string][]string `json:"headers"` // TODO: ensure type matches python lib
	Method       string              `json:"method"`
	Body         string              `json:"body"`
	Ip           string              `json:"ip"`
}

// The responsePayload struct holds the information about the response payload that is passed to the firetail logging endpoint
type ResponsePayload struct {
	StatusCode    int                 `json:"statusCode"`
	ContentLength int                 `json:"contentLength"`
	Body          string              `json:"body"`
	Headers       map[string][]string `json:"headers"` // TODO: ensure type matches python lib
}
