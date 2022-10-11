package firetail

import "fmt"

type ValidationTarget string

const (
	Request  ValidationTarget = "request"
	Response ValidationTarget = "response"
)

type ValidationError struct {
	Target ValidationTarget
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Target, e.Reason)
}

type MethodNotAllowedError struct {
	RequestMethod string
}

func (e *MethodNotAllowedError) Error() string {
	return fmt.Sprintf("%s method is not supported on this route", e.RequestMethod)
}

type RouteNotFoundError struct {
	RequestPath string
}

func (e *RouteNotFoundError) Error() string {
	return fmt.Sprintf("request made to %s but did not match any routes", e.RequestPath)
}
