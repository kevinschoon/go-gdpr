package gdpr

import (
	"bytes"
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
		buf := bytes.NewBuffer(nil)
		sig, err := encodeAndSign(buf, opts.Signer, resp)
		if err != nil {
			return err
		}
		w.Header().Set("X-OpenGDPR-Signature", sig)
		_, err = buf.WriteTo(w)
		return err
	}
}

func postRequest(opts *ServerOptions) Handler {
	validate := ValidateRequest(opts)
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
		req := &Request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			return err
		}
		if err := validate(req); err != nil {
			return err
		}
		buf := bytes.NewBuffer(nil)
		resp, err := opts.Processor.Request(req)
		if err != nil {
			return err
		}
		sig, err := encodeAndSign(buf, opts.Signer, resp)
		if err != nil {
			return err
		}
		w.Header().Set("X-OpenGDPR-Signature", sig)
		w.WriteHeader(201)
		_, err = buf.WriteTo(w)
		return err
	}
}

func deleteRequest(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp, err := opts.Processor.Cancel(p.ByName("id"))
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(nil)
		sig, err := encodeAndSign(buf, opts.Signer, resp)
		if err != nil {
			return err
		}
		w.Header().Set("X-OpenGDPR-Signature", sig)
		_, err = buf.WriteTo(w)
		return err
	}
}

// discovery

func getDiscovery(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
		resp := DiscoveryResponse{
			ApiVersion:                   ApiVersion,
			SupportedSubjectRequestTypes: opts.SubjectTypes,
			SupportedIdentities:          opts.Identities,
			ProcessorCertificate:         opts.ProcessorCertificateUrl,
		}
		return json.NewEncoder(w).Encode(resp)
	}
}

func getCert(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
		if opts.ProcessorCertificateBytes == nil {
			w.WriteHeader(404)
			return nil
		}
		_, err := w.Write(opts.ProcessorCertificateBytes)
		return err
	}
}

// opengdpr_callbacks

func postCallback(opts *ServerOptions) Handler {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
		req := &CallbackRequest{}
		err := json.NewDecoder(r.Body).Decode(req)
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
		// TODO: This is an optional endpoint that
		// can serve the public processor certificate.
		// It's not mentioned in the OpenGDPR spec but
		// is convenient since you don't need to setup
		// a separate file server, etc.
		hm["/cert.pem"] = map[string]Builder{
			"GET": getCert,
		}
	}
	if opts.HandlerMap != nil {
		hm.Merge(opts.HandlerMap)
	}
	return hm
}
