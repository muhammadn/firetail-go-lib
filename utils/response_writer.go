package utils

import "net/http"

// A Firetail ResponseWriter is a responseWriter to which a draft response can be written and later published.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode   int
	ResponseBody []byte
}

func (c *ResponseWriter) WriteHeader(code int) {
	c.StatusCode = code
}

func (c *ResponseWriter) Write(bytes []byte) (int, error) {
	c.ResponseBody = bytes
	return len(bytes), nil
}

func (c *ResponseWriter) Publish() {
	c.ResponseWriter.WriteHeader(c.StatusCode)
	c.ResponseWriter.Write(c.ResponseBody)
}
