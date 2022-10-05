package utils

type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string { return e.Reason }

var RequestValidationError = &ValidationError{"Request failed to validate"}
var ResponseValidationError = &ValidationError{"Response failed to validate"}
