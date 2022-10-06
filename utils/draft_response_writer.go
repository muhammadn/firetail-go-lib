package utils

import "net/http"

// A DraftResponseWriter is a responseWriter to which a draft response can be written and later published.
type DraftResponseWriter struct {
	http.ResponseWriter
	StatusCode   int
	ResponseBody []byte
}

func (c *DraftResponseWriter) WriteHeader(code int) {
	c.StatusCode = code
}

func (c *DraftResponseWriter) Write(bytes []byte) (int, error) {
	c.ResponseBody = bytes
	return len(bytes), nil
}

func (c *DraftResponseWriter) Publish() {
	c.ResponseWriter.WriteHeader(c.StatusCode)
	c.ResponseWriter.Write(c.ResponseBody)
}
