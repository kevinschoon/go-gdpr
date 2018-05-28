package gdpr

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Processor implements the business logic for processing GDPR requests.
// The Processor interface is intended to be wrapped by the Server type
// and provide an HTTP REST server. Any method may return an ErrorResponse
// type which will be serialized as JSON and handled in accordance to
// the OpenGDPR specification.
type Processor interface {
	// Request accepts an incoming Request type
	// and is expected to process it in some way.
	Request(*Request) (*Response, error)
	// Status validates the status of an existing
	// request sent to this processor.
	Status(id string) (*StatusResponse, error)
	// Cancel prevents any further processing of
	// the Request.
	Cancel(id string) (*CancellationResponse, error)
}

// Handler satisfies an incoming HTTP request.
type Handler func(http.ResponseWriter, *http.Request, httprouter.Params) error

// Builder is a functional option to construct a Handler.
type Builder func(opts *ServerOptions) Handler

// HandlerMap is a map of route/methods to Builder.
type HandlerMap map[string]map[string]Builder

// Merge merges another HandlerMap into itself.
func (hm HandlerMap) Merge(other HandlerMap) {
	for key, methods := range other {
		if _, ok := hm[key]; !ok {
			hm[key] = map[string]Builder{}
		}
		for method, builder := range methods {
			hm[key][method] = builder
		}
	}
}

func defaultHandlerMap() HandlerMap {
	return HandlerMap{
		"/opengdpr_requests/:id": map[string]Builder{
			"GET":    getRequest,
			"DELETE": deleteRequest,
		},
		"/opengdpr_requests": map[string]Builder{
			"POST": postRequest,
		},
		"/discovery": map[string]Builder{
			"GET": getDiscovery,
		},
		// TODO: This is an optional endpoint that
		// can serve the public processor certificate.
		// It's not mentioned in the OpenGDPR spec but
		// is convenient since you don't need to setup
		// a separate file server, etc.
		"/cert.pem": map[string]Builder{
			"GET": getCert,
		},
	}
}

// ServerOptions contains configuration options that
// effect the operation of the HTTP server.
type ServerOptions struct {
	Signer                    Signer
	Identities                []Identity
	SubjectTypes              []SubjectType
	Processor                 Processor
	HandlerMap                HandlerMap
	ProcessorDomain           string
	ProcessorCertificate      string
	ProcessorCertificateBytes []byte
}

// Server provides HTTP access to an underlying Processor.
type Server struct {
	handlerFn       func(http.Handler) http.Handler
	router          *httprouter.Router
	processorDomain string
}

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
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
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
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
		resp := DiscoveryResponse{
			ApiVersion:                   ApiVersion,
			SupportedSubjectRequestTypes: opts.SubjectTypes,
			SupportedIdentities:          opts.Identities,
			ProcessorCertificate:         opts.ProcessorCertificate,
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

func (s Server) error(w http.ResponseWriter, err ErrorResponse) {
	w.Header().Set("Cache Control", "no store")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (s Server) handle(fn Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-OpenGDPR-ProcessorDomain", s.processorDomain)
		err := fn(w, r, p)
		if err != nil {
			switch e := err.(type) {
			case ErrorResponse:
				s.error(w, e)
			case *ErrorResponse:
				s.error(w, *e)
			default:
				s.error(w, ErrorResponse{Message: err.Error(), Code: 500})
			}
		}
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.router.ServeHTTP(w, r) }

// NewServer returns a server type that statisfies the
// http.Handler interface.
func NewServer(opts *ServerOptions) *Server {
	server := &Server{
		router:          httprouter.New(),
		processorDomain: opts.ProcessorDomain,
	}
	hm := defaultHandlerMap()
	if opts.HandlerMap != nil {
		hm.Merge(opts.HandlerMap)
	}
	for path, methods := range hm {
		for method, builder := range methods {
			server.router.Handle(method, path, server.handle(builder(opts)))
		}
	}
	return server
}
