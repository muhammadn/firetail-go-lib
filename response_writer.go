package firetail

import "net/http"

type customResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseBody []byte
}

func (c *customResponseWriter) WriteHeader(code int) {
	c.statusCode = code
	c.ResponseWriter.WriteHeader(code)
}

func (c *customResponseWriter) Write(bytes []byte) (int, error) {
	c.responseBody = bytes
	return c.ResponseWriter.Write(bytes)
}
