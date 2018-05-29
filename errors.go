package gdpr

import (
	"fmt"
	"net/http"
)

// ErrNotFound indicates a request could
// not be found by the processor.
func ErrNotFound(id string) error {
	return ErrorResponse{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("request %s not found", id),
	}
}

// ErrUnsupportedRequestType indicates the processor
// cannot fullfil a request for the given RequestType.
func ErrUnsupportedRequestType(st SubjectType) error {
	return ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: fmt.Sprintf("unsupported request type: %s", st),
	}
}

// ErrUnsupportedIdentity indicates the processor
// does not support the given identity type.
func ErrUnsupportedIdentity(id Identity) error {
	return ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: fmt.Sprintf("unsupported identity: %s/%s", id.Type, id.Format),
	}
}

// ErrMissingRequiredField indicates the request
// is missing a required field.
func ErrMissingRequiredField(field string) error {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("missing required field: %s", field),
	}
}

// ErrInvalidRequestSignature indicates the payload could not
// be verified with the given signature.
func ErrInvalidRequestSignature(signature string, err error) error {
	return ErrorResponse{
		Code:    http.StatusForbidden,
		Message: fmt.Sprintf("could not validate request signature: %s", signature),
		Errors:  []Error{Error{Message: err.Error()}},
	}
}
