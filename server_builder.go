package gdpr

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// opengdpr_requests

func getRequest(s Server, g Gdpr) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp, err := g.Status(p.ByName("id"))
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

func postRequest(s Server, g Gdpr) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		req := &Request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			return err
		}
		resp, err := g.Request(*req)
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

func deleteRequest(s Server, g Gdpr) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp, err := g.Cancel(p.ByName("id"))
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

// discovery

func getDiscovery(s Server, g Gdpr) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp := DiscoveryResponse{
			ApiVersion:                   ApiVersion,
			SupportedSubjectRequestTypes: s.subjectTypes,
			SupportedIdentities:          s.identities,
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

// callback

func postCallback(s Server, g Gdpr) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		req := CallbackRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}
		return g.Callback(req)
	}
}
