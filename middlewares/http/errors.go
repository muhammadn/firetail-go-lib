package firetail

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// ErrorAppspecNotFound is used at initialisation/startup when the OpenAPI appspec file the library has been configured to use cannot be found
type ErrorAppspecNotFound struct {
	ProvidedPath string // The path that was used to attempt to load the appspec
}

func (e ErrorAppspecNotFound) Error() string {
	// TODO
	return "ErrorAppspecNotFound"
}

// ErrorAppspecInvalid is used at initialisation/startup when the OpenAPI appspec file is malformed
type ErrorAppspecInvalid struct {
	Reason error // The error that occurred during initialisation of the middleware due to the appspec being invalid
}

func (e ErrorAppspecInvalid) Error() string {
	// TODO
	return "ErrorAppspecInvalid"
}

// ErrorAtRequest is an interface that extends the standard error interface for errors that occur during the handling of a request.
// To satisfy this interface, errors should implement a method which returns an appropriate HTTP status code to provide the client.
type ErrorAtRequest interface {
	error

	// Should return the appropriate HTTP status code to provide in response to the request for which the error occured
	StatusCode() int
}

// ErrorRouteNotFound is used when a request is made for which no corresponding route in the OpenAPI spec could be found
type ErrorRouteNotFound struct {
	RequestedPath string // The path that was requested for which no corresponding route could be found
}

func (e ErrorRouteNotFound) StatusCode() int {
	return 404
}

func (e ErrorRouteNotFound) Error() string {
	return fmt.Sprintf("no matching route found for request path %s", e.RequestedPath)
}

// ErrorUnsupportedMethod is used when a request is made which corresponds to a route in the OpenAPI spec, but that route
// doesn't support the HTTP method with which the request was made
type ErrorUnsupportedMethod struct {
	RequestedRoute  string // The route that corresponds to the path that was requested for which the method is not supported
	RequestedMethod string // The method which was requested but is not supported on the route corresponding to the request path
}

func (e ErrorUnsupportedMethod) StatusCode() int {
	return 405
}

func (e ErrorUnsupportedMethod) Error() string {
	return fmt.Sprintf("%s route does not support %s method", e.RequestedRoute, e.RequestedMethod)
}

// ErrorRequestHeadersInvalid is used when any of the headers of a request don't conform to the schema in the OpenAPI spec, except
// for the Content-Type header for which an ErrorRequestContentTypeInvalid is used
type ErrorRequestHeadersInvalid struct {
}

func (e ErrorRequestHeadersInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestHeadersInvalid) Error() string {
	// TODO
	return "ErrorRequestHeadersInvalid"
}

// ErrorRequestHeadersInvalid is used when the Content-Type header of a request doesn't conform to the schema in the OpenAPI spec
type ErrorRequestContentTypeInvalid struct {
	RequestedContentType string
	RequestedRoute       string
}

func (e ErrorRequestContentTypeInvalid) StatusCode() int {
	return 415
}

func (e ErrorRequestContentTypeInvalid) Error() string {
	return fmt.Sprintf("%s route does not support %s method", e.RequestedRoute, e.RequestedContentType)
}

// ErrorRequestQueryParamsInvalid is used when the query params of a request don't conform to the schema in the OpenAPI spec
type ErrorRequestQueryParamsInvalid struct {
}

func (e ErrorRequestQueryParamsInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestQueryParamsInvalid) Error() string {
	// TODO
	return "ErrorRequestQueryParamsInvalid"
}

// ErrorRequestPathParamsInvalid is used when the path params of a request don't conform to the schema in the OpenAPI spec
type ErrorRequestPathParamsInvalid struct {
}

func (e ErrorRequestPathParamsInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestPathParamsInvalid) Error() string {
	// TODO
	return "ErrorRequestPathParamsInvalid"
}

// ErrorRequestBodyInvalid is used when the body of a request doesn't conform to the schema in the OpenAPI spec
type ErrorRequestBodyInvalid struct {
	Reason string
}

func (e ErrorRequestBodyInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestBodyInvalid) Error() string {
	return fmt.Sprintf("response body invalid, reason: %s", e.Reason)
}

// ErrorAuthNoMatchingSchema is used when a request doesn't satisfy any of the securitySchemes corresponding to the route that the request matched in the OpenAPI spec
type ErrorAuthNoMatchingSchema struct {
	SecurityRequirements openapi3.SecurityRequirements
}

func (e ErrorAuthNoMatchingSchema) StatusCode() int {
	return 401
}

func (e ErrorAuthNoMatchingSchema) Error() string {
	return fmt.Sprintf("request did not satisfy any of the following security requirements: %v", e.SecurityRequirements)
}

// ErrorResponseHeadersInvalid is used when any of the headers of a response don't conform to the schema in the OpenAPI spec
type ErrorResponseHeadersInvalid struct {
}

func (e ErrorResponseHeadersInvalid) StatusCode() int {
	return 500
}

func (e ErrorResponseHeadersInvalid) Error() string {
	// TODO
	return "ErrorResponseHeadersInvalid"
}

// ErrorResponseHeadersInvalid is used when the body of a response doesn't conform to the schema in the OpenAPI spec
type ErrorResponseBodyInvalid struct {
	Reason string
}

func (e ErrorResponseBodyInvalid) StatusCode() int {
	return 500
}

func (e ErrorResponseBodyInvalid) Error() string {
	return fmt.Sprintf("response body invalid, reason: %s", e.Reason)
}

// ErrorResponseStatusCodeInvalid is used when the status code of a response doesn't conform to the schema in the OpenAPI spec
type ErrorResponseStatusCodeInvalid struct {
}

func (e ErrorResponseStatusCodeInvalid) StatusCode() int {
	return 500
}

func (e ErrorResponseStatusCodeInvalid) Error() string {
	// TODO
	return "ErrorResponseStatusCodeInvalid"
}
