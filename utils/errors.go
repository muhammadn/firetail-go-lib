package utils

import "github.com/getkin/kin-openapi/routers"

type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string { return e.Reason }

var ErrRequestValidationFailed = &ValidationError{"Request failed to validate"}
var ErrResponseValidationFailed = &ValidationError{"Response failed to validate"}
var ErrMethodNotAllowed = routers.ErrMethodNotAllowed
var ErrPathNotFound = routers.ErrPathNotFound
