package gdpr

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// opengdpr_requests

func getRequest(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp, err := opts.Processor.Status(p.ByName("id"))
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

func postRequest(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		req := &Request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			return err
		}
		resp, err := opts.Processor.Request(req)
		if err != nil {
			return err
		}
		w.Header().Add("X-OpenGDPR-Signature", req.Signature())
		return json.NewEncoder(w).Encode(resp)
	}
}

func deleteRequest(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp, err := opts.Processor.Cancel(p.ByName("id"))
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

// discovery

func getDiscovery(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp := DiscoveryResponse{
			ApiVersion:                   ApiVersion,
			SupportedSubjectRequestTypes: opts.SubjectTypes,
			SupportedIdentities:          opts.Identities,
		}
		return json.NewEncoder(w).Encode(resp)
	}
}
