package gdpr

import (
	"bytes"
	"encoding/json"
)

// SupportedFunc returns a function that checks if the server can
// support a specific request.
func SupportedFunc(opts *ServerOptions) func(*Request) error {
	subjectMap := map[SubjectType]bool{}
	for _, subjectType := range opts.SubjectTypes {
		subjectMap[subjectType] = true
	}
	identityMap := map[string]bool{}
	for _, identity := range opts.Identities {
		identityMap[string(identity.Type)+string(identity.Format)] = true
	}
	return func(req *Request) error {
		if _, ok := subjectMap[req.SubjectRequestType]; !ok {
			return ErrUnsupportedRequestType(req.SubjectRequestType)
		}
		for _, identity := range req.SubjectIdentities {
			if _, ok := identityMap[string(identity.Type)+string(identity.Format)]; !ok {
				return ErrUnsupportedIdentity(identity)
			}
		}
		return nil
	}
}

func ValidateRequest(opts *ServerOptions) func(*Request) error {
	fn := SupportedFunc(opts)
	return func(req *Request) error {
		if req.SubjectRequestId == "" {
			return ErrMissingRequiredField("subject_request_id")
		}
		if len(req.SubjectIdentities) == 0 {
			return ErrMissingRequiredField("subject_identities")
		}
		return fn(req)
	}
}

// encodeAndSign encodes some value and generates a signature
func encodeAndSign(buf *bytes.Buffer, s Signer, v interface{}) (string, error) {
	err := json.NewEncoder(buf).Encode(v)
	if err != nil {
		return "", err
	}
	sig, err := s.Sign(buf.Bytes())
	if err != nil {
		return "", err
	}
	return sig, nil
}
