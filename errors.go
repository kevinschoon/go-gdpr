package gdpr

import (
	"fmt"
)

// NotFound indicates no request exists
// for the given ID.
func NotFound(id string) error {
	return ErrorResponse{
		Code:    404,
		Message: fmt.Sprintf("request %s not found", id),
	}
}
