package gdpr

import (
	"encoding/json"
	"io"

	"github.com/julienschmidt/httprouter"
)

// opengdpr_requests

func getRequest(opts *ServerOptions) Handler {
	return func(w io.Writer, _ io.Reader, p httprouter.Params) error {
		resp, err := opts.Processor.Status(p.ByName("id"))
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

func postRequest(opts *ServerOptions) Handler {
	validate := ValidateRequest(opts)
	return func(w io.Writer, r io.Reader, _ httprouter.Params) error {
		req := &Request{}
		err := json.NewDecoder(r).Decode(req)
		if err != nil {
			return err
		}
		if err := validate(req); err != nil {
			return err
		}
		resp, err := opts.Processor.Request(req)
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

func deleteRequest(opts *ServerOptions) Handler {
	return func(w io.Writer, _ io.Reader, p httprouter.Params) error {
		resp, err := opts.Processor.Cancel(p.ByName("id"))
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

// discovery

func getDiscovery(opts *ServerOptions) Handler {
	return func(w io.Writer, _ io.Reader, _ httprouter.Params) error {
		resp := DiscoveryResponse{
			ApiVersion:                   ApiVersion,
			SupportedSubjectRequestTypes: opts.SubjectTypes,
			SupportedIdentities:          opts.Identities,
			ProcessorCertificate:         opts.ProcessorCertificateUrl,
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

// opengdpr_callbacks

func postCallback(opts *ServerOptions) Handler {
	return func(_ io.Writer, r io.Reader, _ httprouter.Params) error {
		req := &CallbackRequest{}
		err := json.NewDecoder(r).Decode(req)
		if err != nil {
			return err
		}
		return opts.Controller.Callback(req)
	}
}

func buildHandlerMap(opts *ServerOptions) HandlerMap {
	hm := HandlerMap{}
	// controller map
	if hasController(opts) {
		hm["/opengdpr_callbacks"] = map[string]Builder{
			"POST": postCallback,
		}
	}
	if hasProcessor(opts) {
		hm["/opengdpr_requests/:id"] = map[string]Builder{
			"GET":    getRequest,
			"DELETE": deleteRequest,
		}
		hm["/opengdpr_requests"] = map[string]Builder{
			"POST": postRequest,
		}
		hm["/discovery"] = map[string]Builder{
			"GET": getDiscovery,
		}
	}
	if opts.HandlerMap != nil {
		hm.Merge(opts.HandlerMap)
	}
	return hm
}
