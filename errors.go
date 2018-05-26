package gdpr

import (
	"fmt"
)

func ErrNotFound(id string) error {
	return ErrorResponse{
		Code:    404,
		Message: fmt.Sprintf("request %s not found", id),
	}
}

func ErrUnsupportedRequestType(st SubjectType) error {
	return ErrorResponse{
		Code:    501,
		Message: fmt.Sprintf("unsupported request type: %s", st),
	}
}

func ErrUnsupportedIdentity(id Identity) error {
	return ErrorResponse{
		Code:    501,
		Message: fmt.Sprintf("unsupported identity: %s/%s", id.Type, id.Format),
	}
}

func ErrMissingRequiredField(field string) error {
	return ErrorResponse{
		Code:    400,
		Message: fmt.Sprintf("missing required field: %s", field),
	}
}
