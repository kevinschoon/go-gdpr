package gdpr

// SupportedFunc returns a function that checks if the server can
// support a specific request.
func SupportedFunc(opts *ServerOptions) func(*Request) bool {
	subjectMap := map[SubjectType]bool{}
	for _, subjectType := range opts.SubjectTypes {
		subjectMap[subjectType] = true
	}
	identityMap := map[string]bool{}
	for _, identity := range opts.Identities {
		identityMap[string(identity.Type)+string(identity.Format)] = true
	}
	return func(req *Request) bool {
		if _, ok := subjectMap[req.SubjectRequestType]; !ok {
			return false
		}
		for _, identity := range req.SubjectIdentities {
			if _, ok := identityMap[string(identity.Type)+string(identity.Format)]; !ok {
				return false
			}
		}
		return true
	}
}
