package firetail

// The Payload struct holds a payload to be sent to the firetail logging endpoint
type LoggingPayload struct {
	Version        string          `json:"version"`
	DateCreated    int64           `json:"dateCreated"`    // UNIX Milliseconds
	Execution_time int64           `json:"execution_time"` // Milliseconds
	Source_code    string          `json:"source_code"`
	Req            RequestPayload  `json:"req"`
	Resp           ResponsePayload `json:"resp"`
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
	Status_code int                 `json:"status_code"`
	Content_len int                 `json:"content_len"`
	Content_enc string              `json:"content_enc"`
	Body        string              `json:"body"`
	Headers     map[string][]string `json:"headers"` // TODO: ensure type matches python lib
}
