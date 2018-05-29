package gdpr

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Controller makes new requests to a Processor
// and processes Callback requests.
type Controller interface {
	// Process a callback from a remote
	// processor.
	Callback(req *CallbackRequest) error
}

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

// ServerOptions contains configuration options that
// effect the operation of the HTTP server.
type ServerOptions struct {
	// Controller to process callback requests.
	Controller Controller
	// Processor to handle GDPR requests.
	Processor Processor
	// Signs all responses.
	Signer Signer
	// Verifies any incoming callbacks.
	Verifier Verifier
	// Array of identity types supported by
	// the server.
	Identities []Identity
	// Array of subject types supported by
	// the server.
	SubjectTypes []SubjectType
	// Optional map allowing the user to
	// override specific routes.
	HandlerMap HandlerMap
	// Processor domain of this server.
	ProcessorDomain string
	// Optional public certificate to serve to
	// any clients for response verification.
	ProcessorCertificateBytes []byte
	// Remote URL where the public certificate
	// of this server can be downloaded and used
	// to verify subsequent response payload
	ProcessorCertificateUrl string
}

// Server exposes an HTTP interface to an underlying
// Processor or Controller. Server relies on httprouter
// for route matching, in the future we might expand
// this to support other mux and frameworks.
type Server struct {
	headers         http.Header
	handlerFn       func(http.Handler) http.Handler
	router          *httprouter.Router
	processorDomain string
}

func (s *Server) setHeaders(w http.ResponseWriter) {
	for key, headers := range s.headers {
		for _, header := range headers {
			w.Header().Add(key, header)
		}
	}
}

func (s *Server) error(w http.ResponseWriter, err ErrorResponse) {
	w.Header().Set("Cache Control", "no-store")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (s *Server) handle(fn Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.setHeaders(w)
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
		headers:         http.Header{},
		processorDomain: opts.ProcessorDomain,
	}
	server.headers.Set("Accept", "application/json")
	server.headers.Set("Content-Type", "application/json")
	if hasProcessor(opts) {
		server.headers.Set("X-OpenGDPR-ProcessorDomain", opts.ProcessorDomain)
	}
	router := httprouter.New()
	hm := buildHandlerMap(opts)
	for path, methods := range hm {
		for method, builder := range methods {
			router.Handle(method, path, server.handle(builder(opts)))
		}
	}
	server.router = router
	return server
}

func hasController(opts *ServerOptions) bool {
	return opts.Controller != nil
}

func hasProcessor(opts *ServerOptions) bool {
	return opts.Processor != nil
}
