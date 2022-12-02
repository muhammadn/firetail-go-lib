package firetail

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3filter"
)

// InvalidConfiguration is used by middleware constructors if the configuration they are provided is invalid for some reason
type ErrorInvalidConfiguration struct {
	Err error // The path that was used to attempt to load the appspec
}

func (e ErrorInvalidConfiguration) Error() string {
	return fmt.Sprintf("invalid configuration: %s", e.Err.Error())
}

// ErrorAppspecInvalid is used at initialisation/startup when the OpenAPI appspec file is malformed
type ErrorAppspecInvalid struct {
	Err error // The error that occurred during initialisation of the middleware due to the appspec being invalid
}

func (e ErrorAppspecInvalid) Error() string {
	return fmt.Sprintf("invalid appspec: %s", e.Err.Error())
}

// ErrorAtRequest is an interface that extends the standard error interface for errors that occur during the handling of a request.
// To satisfy this interface, errors should implement a method which returns an appropriate HTTP status code to provide the client.
type ErrorAtRequest interface {
	error

	// Should return the appropriate HTTP status code to provide in response to the request for which the error occured
	StatusCode() int

	// Should return a title string appropriate for public RFC7807 error responses
	Title() string
}

// ErrorAtRequestUnspecified is used to wrap errors that are returned at request time, but aren't able to be broken down into more useful information
type ErrorAtRequestUnspecified struct {
	Err error
}

func (e ErrorAtRequestUnspecified) StatusCode() int {
	return 500
}

func (e ErrorAtRequestUnspecified) Title() string {
	return "internal server error"
}

func (e ErrorAtRequestUnspecified) Error() string {
	return e.Err.Error()
}

// ErrorRouteNotFound is used when a request is made for which no corresponding route in the OpenAPI spec could be found
type ErrorRouteNotFound struct {
	RequestedPath string // The path that was requested for which no corresponding route could be found
}

func (e ErrorRouteNotFound) StatusCode() int {
	return 404
}

func (e ErrorRouteNotFound) Title() string {
	return fmt.Sprintf("the resource \"%s\" could not be found", e.RequestedPath)
}

func (e ErrorRouteNotFound) Error() string {
	return fmt.Sprintf("a path for \"%s\" could not be found in your appspec", e.RequestedPath)
}

// ErrorUnsupportedMethod is used when a request is made which corresponds to a route in the OpenAPI spec, but that route
// doesn't support the HTTP method with which the request was made
type ErrorUnsupportedMethod struct {
	RequestedPath   string // The route that corresponds to the path that was requested for which the method is not supported
	RequestedMethod string // The method which was requested but is not supported on the route corresponding to the request path
}

func (e ErrorUnsupportedMethod) StatusCode() int {
	return 405
}

func (e ErrorUnsupportedMethod) Title() string {
	return fmt.Sprintf("the resource \"%s\" does not support the \"%s\" method", e.RequestedPath, e.RequestedMethod)
}

func (e ErrorUnsupportedMethod) Error() string {
	return fmt.Sprintf("the path for \"%s\" in your appspec does not support the method \"%s\"", e.RequestedPath, e.RequestedMethod)
}

// ErrorRequestHeadersInvalid is used when any of the headers of a request don't conform to the schema in the OpenAPI spec, except
// for the Content-Type header for which an ErrorRequestContentTypeInvalid is used
type ErrorRequestHeadersInvalid struct {
	Err error
}

func (e ErrorRequestHeadersInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestHeadersInvalid) Title() string {
	return "something's wrong with your request headers"
}

func (e ErrorRequestHeadersInvalid) Error() string {
	return fmt.Sprintf("the request's headers did not match your appspec: %s", e.Err.Error())
}

// ErrorRequestHeadersInvalid is used when the Content-Type header of a request doesn't conform to the schema in the OpenAPI spec
type ErrorRequestContentTypeInvalid struct {
	RequestedContentType string
	RequestedRoute       string
}

func (e ErrorRequestContentTypeInvalid) StatusCode() int {
	return 415
}

func (e ErrorRequestContentTypeInvalid) Title() string {
	return fmt.Sprintf("the path for \"%s\" in your appspec does not support the content type \"%s\"", e.RequestedRoute, e.RequestedContentType)
}

func (e ErrorRequestContentTypeInvalid) Error() string {
	return fmt.Sprintf("the path for \"%s\" in your appspec does not support content type \"%s\"", e.RequestedRoute, e.RequestedContentType)
}

// ErrorRequestQueryParamsInvalid is used when the query params of a request don't conform to the schema in the OpenAPI spec
type ErrorRequestQueryParamsInvalid struct {
	Err error
}

func (e ErrorRequestQueryParamsInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestQueryParamsInvalid) Title() string {
	return "something's wrong with your query parameters"
}

func (e ErrorRequestQueryParamsInvalid) Error() string {
	return fmt.Sprintf("the request's query parameters did not match your appspec: %s", e.Err.Error())
}

// ErrorRequestPathParamsInvalid is used when the path params of a request don't conform to the schema in the OpenAPI spec
type ErrorRequestPathParamsInvalid struct {
	Err error
}

func (e ErrorRequestPathParamsInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestPathParamsInvalid) Title() string {
	return "something's wrong with your path parameters"
}

func (e ErrorRequestPathParamsInvalid) Error() string {
	return fmt.Sprintf("the request's path parameters did not match your appspec: %s", e.Err.Error())
}

// ErrorRequestBodyInvalid is used when the body of a request doesn't conform to the schema in the OpenAPI spec
type ErrorRequestBodyInvalid struct {
	Err error
}

func (e ErrorRequestBodyInvalid) StatusCode() int {
	return 400
}

func (e ErrorRequestBodyInvalid) Title() string {
	return "something's wrong with your request body"
}

func (e ErrorRequestBodyInvalid) Error() string {
	return fmt.Sprintf("the request's body did not match your appspec: %s", e.Err.Error())
}

// ErrorAuthNoMatchingSchema is used when a request doesn't satisfy any of the securitySchemes corresponding to the route that the request matched in the OpenAPI spec
type ErrorAuthNoMatchingScheme struct {
	Err *openapi3filter.SecurityRequirementsError
}

func (e ErrorAuthNoMatchingScheme) StatusCode() int {
	return 401
}

func (e ErrorAuthNoMatchingScheme) Title() string {
	return "you're not authorized to do this"
}

func (e ErrorAuthNoMatchingScheme) Error() string {
	errString := fmt.Sprintf("the request did not satisfy the security requirements in your appspec: %s, errors: ", e.Err.Error())
	for i, err := range e.Err.Errors {
		errString += err.Error()
		if i < len(e.Err.Errors)-1 {
			errString += ", "
		}
	}
	return errString
}

// ErrorAuthSchemaNotImplemented is used when a request is made to a path that has a security scheme requirement that has not been implemented in the application
type ErrorAuthSchemeNotImplemented struct {
	MissingScheme string
}

func (e ErrorAuthSchemeNotImplemented) Error() string {
	return fmt.Sprintf("the security scheme \"%s\" from your appspec has not been implemented in the application", e.MissingScheme)
}

// ErrorResponseHeadersInvalid is used when any of the headers of a response don't conform to the schema in the OpenAPI spec
// Currently not implemented as the underlying kin-openapi module doesn't perform response header validation.
// See the open issue here: https://github.com/getkin/kin-openapi/issues/201
// TODO: Open source contribution to kin-openapi?
type ErrorResponseHeadersInvalid struct {
	Err error
}

func (e ErrorResponseHeadersInvalid) StatusCode() int {
	return 500
}

func (e ErrorResponseHeadersInvalid) Title() string {
	return "internal server error"
}

func (e ErrorResponseHeadersInvalid) Error() string {
	return fmt.Sprintf("the response's headers did not match your appspec: %s", e.Err.Error())
}

// ErrorResponseHeadersInvalid is used when the body of a response doesn't conform to the schema in the OpenAPI spec
type ErrorResponseBodyInvalid struct {
	Err error
}

func (e ErrorResponseBodyInvalid) StatusCode() int {
	return 500
}

func (e ErrorResponseBodyInvalid) Title() string {
	return "internal server error"
}

func (e ErrorResponseBodyInvalid) Error() string {
	return fmt.Sprintf("the response's body did not match your appspec: %s", e.Err.Error())
}

// ErrorResponseStatusCodeInvalid is used when the status code of a response doesn't conform to the schema in the OpenAPI spec
type ErrorResponseStatusCodeInvalid struct {
	RespondedStatusCode int
}

func (e ErrorResponseStatusCodeInvalid) StatusCode() int {
	return 500
}

func (e ErrorResponseStatusCodeInvalid) Title() string {
	return "internal server error"
}

func (e ErrorResponseStatusCodeInvalid) Error() string {
	return fmt.Sprintf("the response's status code did not match your appspec: %d", e.RespondedStatusCode)
}
