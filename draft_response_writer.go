package firetail

import "net/http"

// A draftResponseWriter is a responseWriter to which a draft response can be written and later published.
type draftResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseBody []byte
}

func (c *draftResponseWriter) WriteHeader(code int) {
	c.statusCode = code
}

func (c *draftResponseWriter) Write(bytes []byte) (int, error) {
	c.responseBody = bytes
	return len(bytes), nil
}

func (c *draftResponseWriter) Publish() {
	c.ResponseWriter.WriteHeader(c.statusCode)
	c.ResponseWriter.Write(c.responseBody)
}
