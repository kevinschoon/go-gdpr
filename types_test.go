package gdpr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorEncoding(t *testing.T) {
	e := &ErrorResponse{
		Code:    400,
		Message: "Oh No!",
		Errors: []Error{
			Error{
				Reason:  "something broke",
				Domain:  "some-domain.com",
				Message: "it's broken",
			},
		},
	}
	raw, err := json.Marshal(e)
	assert.NoError(t, err)
	t.Log(string(raw))
	expected := []byte(`{"error":{"code":400,"message":"Oh No!","errors":[{"domain":"some-domain.com","reason":"something broke","message":"it's broken"}]}}`)
	assert.Equal(t, string(raw), string(expected))
}

func TestErrorDecoding(t *testing.T) {
	raw := []byte(`{"error":{"code":400,"message":"Oh No!","errors":[{"domain":"some-domain.com","reason":"something broke","message":"it's broken"}]}}`)
	e := &ErrorResponse{}
	assert.NoError(t, json.Unmarshal(raw, e))
	t.Log(e)
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, "Oh No!", e.Message)
	assert.Equal(t, "something broke", e.Errors[0].Reason)
}
