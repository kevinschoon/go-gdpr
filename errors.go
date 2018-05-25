package gdpr

import (
	"fmt"
)

func NotFound(id string) error {
	return ErrorResponse{
		Code:    404,
		Message: fmt.Sprintf("request %s not found", id),
	}
}
