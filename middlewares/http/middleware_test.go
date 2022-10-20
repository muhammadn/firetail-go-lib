package firetail

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/sbabiv/xml2map"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var healthHandler http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{\"description\":\"test description\"}"))
})

var healthHandlerWithWrongResponseBody http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{\"description\":\"another test description\"}"))
})

func TestValidRequestAndResponse(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"test description\"}"))),
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
	assert.Equal(t, "{\"description\":\"test description\"}", string(respBody))
}

func TestInvalidSpecPath(t *testing.T) {
	_, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec-not-here.yaml",
	})
	require.IsType(t, ErrorInvalidConfiguration{}, err)
	require.Equal(t, "invalid configuration, err: open ./test-spec-not-here.yaml: no such file or directory", err.Error())
}

func TestInvalidSpec(t *testing.T) {
	_, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec-invalid.yaml",
	})
	require.Equal(t, "invalid paths: a short description of the response is required", err.Error())
}

func TestRequestToInvalidRoute(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
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
	assert.Equal(t, "no matching route found for request path /not-implemented", string(respBody))
}

func TestRequestWithDisallowedMethod(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
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
	assert.Equal(t, "/implemented does not support GET method", string(respBody))
}

func TestRequestWithInvalidBody(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
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
	assert.Equal(
		t,
		"request body has an error: doesn't match the schema: Error at \"/description\": property \"description\" is missing\nSchema:\n  {\n    \"additionalProperties\": false,\n    \"properties\": {\n      \"description\": {\n        \"enum\": [\n          \"test description\"\n        ],\n        \"type\": \"string\"\n      }\n    },\n    \"required\": [\n      \"description\"\n    ],\n    \"type\": \"object\"\n  }\n\nValue:\n  {}\n",
		string(respBody),
	)
}

func TestInvalidResponseBody(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandlerWithWrongResponseBody)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"test description\"}"))),
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
		"response body doesn't match the schema: Error at \"/description\": value is not one of the allowed values\nSchema:\n  {\n    \"enum\": [\n      \"test description\"\n    ],\n    \"type\": \"string\"\n  }\n\nValue:\n  \"another test description\"\n",
		string(respBody),
	)
}

func TestDisabledRequestValidation(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath:   "./test-spec.yaml",
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
	assert.Equal(t, "{\"description\":\"test description\"}", string(respBody))
}

func TestDisabledResponseValidation(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath:   "./test-spec.yaml",
		DisableValidation: true,
	})
	require.Nil(t, err)
	handler := middleware(healthHandlerWithWrongResponseBody)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"test description\"}"))),
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
	assert.Equal(t, "{\"description\":\"another test description\"}", string(respBody))
}

func TestUnexpectedContentType(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("{\"description\":\"test description\"}"))),
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
	assert.Equal(t, "/implemented route does not support content type text/plain", string(respBody))
}

func TestCustomXMLDecoder(t *testing.T) {
	middleware, err := GetMiddleware(&Options{
		OpenapiSpecPath: "./test-spec.yaml",
		CustomBodyDecoders: map[string]openapi3filter.BodyDecoder{
			"application/xml": func(r io.Reader, h http.Header, sr *openapi3.SchemaRef, ef openapi3filter.EncodingFn) (interface{}, error) {
				decoder := xml2map.NewDecoder(r)
				result, err := decoder.Decode()
				if err != nil {
					return nil, err
				}
				log.Println(result)
				return result, nil
			},
		},
	})
	require.Nil(t, err)
	handler := middleware(healthHandler)
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(
		"POST", "/implemented",
		io.NopCloser(bytes.NewBuffer([]byte("<description>test description</description>"))),
	)
	request.Header.Add("Content-Type", "application/xml")
	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, 200, responseRecorder.Code)

	require.Contains(t, responseRecorder.HeaderMap, "Content-Type")
	require.GreaterOrEqual(t, len(responseRecorder.HeaderMap["Content-Type"]), 1)
	assert.Len(t, responseRecorder.HeaderMap["Content-Type"], 1)
	assert.Equal(t, "application/json", responseRecorder.HeaderMap["Content-Type"][0])

	respBody, err := io.ReadAll(responseRecorder.Body)
	require.Nil(t, err)
	assert.Equal(t, "{\"description\":\"test description\"}", string(respBody))
}
