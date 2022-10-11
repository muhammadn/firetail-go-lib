package firetail

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var healthHandler http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("I'm Healthy! :)"))
})

func TestValidRequest(t *testing.T) {
	// Get a middleware
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)

	// Create our handler
	handler := middleware(healthHandler)

	// Create our response recorder & test request
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/health", nil)

	// Handle the test request & record the response
	handler.ServeHTTP(responseRecorder, request)

	// Status code should be 200
	assert.Equal(t, 200, responseRecorder.Code)

	// Should have a Content-Type: text/plain header
	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, responseRecorder.HeaderMap["Content-Type"][0], "text/plain")

	// Response body should be "I'm Healthy! :)"
	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "I'm Healthy! :)", string(respBody))
}
