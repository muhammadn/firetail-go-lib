package firetail

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var healthHandler http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{\"description\":\"A test JSON object\"}"))
})

var healthHandlerWithWrongResponseBody http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{\"description\":\"A different test JSON object\"}"))
})

func TestValidRequestAndResponse(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"A test JSON object\"}"))),
	)
	request.Header.Add("Content-Type", "application/json")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 200, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "application/json", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "{\"description\":\"A test JSON object\"}", string(respBody))
}

func TestInvalidSpecPath(t *testing.T) {
	_, err := GetMiddleware(&Options{
		SpecPath: "./test-spec-not-here.yaml",
	})
	require.IsType(t, &fs.PathError{}, err)
}

func TestInvalidSpec(t *testing.T) {
	_, err := GetMiddleware(&Options{
		SpecPath: "./test-spec-invalid.yaml",
	})
	require.Equal(t, "invalid paths: a short description of the response is required", err.Error())
}

func TestRequestToInvalidRoute(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest("GET", "/not-implemented", nil)
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 404, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "text/plain", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "404 (Not Found): request made to /not-implemented but did not match any routes", string(respBody))
}

func TestRequestWithDisallowedMethod(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest("GET", "/implemented", nil)
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 405, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "text/plain", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "405 (Method Not Allowed): GET method is not supported on this route", string(respBody))
}

func TestRequestWithInvalidBody(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{}"))),
	)
	request.Header.Add("Content-Type", "application/json")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 400, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "text/plain", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "400 (Bad Request): validation failed on request: request body has an error: doesn't match the schema: Error at \"/description\": property \"description\" is missing\nSchema:\n  {\n    \"additionalProperties\": false,\n    \"properties\": {\n      \"description\": {\n        \"enum\": [\n          \"A test JSON object\"\n        ],\n        \"type\": \"string\"\n      }\n    },\n    \"required\": [\n      \"description\"\n    ],\n    \"type\": \"object\"\n  }\n\nValue:\n  {}\n", string(respBody))
}

func TestInvalidResponseBody(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandlerWithWrongResponseBody)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"A test JSON object\"}"))),
	)
	request.Header.Add("Content-Type", "application/json")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 500, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "text/plain", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(
		t,
		"500 (Internal Server Error): validation failed on response: response body doesn't match the schema: Error at \"/description\": value is not one of the allowed values\nSchema:\n  {\n    \"enum\": [\n      \"A test JSON object\"\n    ],\n    \"type\": \"string\"\n  }\n\nValue:\n  \"A different test JSON object\"\n",
		string(respBody),
	)
}

func TestDisabledRequestValidation(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath:          "./test-spec.yaml",
		DisableValidation: true,
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"Another test JSON object\"}"))),
	)
	request.Header.Add("Content-Type", "application/json")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 200, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "application/json", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "{\"description\":\"A test JSON object\"}", string(respBody))
}

func TestDisabledResponseValidation(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath:          "./test-spec.yaml",
		DisableValidation: true,
	})
	require.Nil(t, err)
	handler := middleware(healthHandlerWithWrongResponseBody)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"A test JSON object\"}"))),
	)
	request.Header.Add("Content-Type", "application/json")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 200, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "application/json", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "{\"description\":\"A different test JSON object\"}", string(respBody))
}

func TestUnexpectedContentType(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		SpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"A test JSON object\"}"))),
	)
	request.Header.Add("Content-Type", "text/plain")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 415, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "text/plain", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "415 (Unsupported Media Type): content type 'text/plain' is not supported on this route", string(respBody))
}
