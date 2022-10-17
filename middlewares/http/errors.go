package firetail

import "fmt"

type ValidationTarget string

const (
	Request  ValidationTarget = "request"
	Response ValidationTarget = "response"
)

// A ValidationError should return a 400 error to the client if the Request failed to validate, and a
// 500 error to the client if the Response failed to validate
type ValidationError struct {
	Target ValidationTarget
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Target, e.Reason)
}

// A ContentTypeError should return a 415 error to the client
type ContentTypeError struct {
	Actual string
}

func (e *ContentTypeError) Error() string {
	return fmt.Sprintf("content type '%s' is not supported on this route", e.Actual)
}

// A MethodNotAllowedError should return a 405 error to the client
type MethodNotAllowedError struct {
	RequestMethod string
}

func (e *MethodNotAllowedError) Error() string {
	return fmt.Sprintf("%s method is not supported on this route", e.RequestMethod)
}

// A RouteNotFoundError should return a 404 error to the client
type RouteNotFoundError struct {
	RequestPath string
}

func (e *RouteNotFoundError) Error() string {
	return fmt.Sprintf("request made to %s but did not match any routes", e.RequestPath)
}
