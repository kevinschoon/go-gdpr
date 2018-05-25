package gdpr

import (
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
	}
}

// ServerOptions contains configuration options that
// effect the operation of the HTTP server.
type ServerOptions struct {
	Identities      []Identity
	SubjectTypes    []SubjectType
	Processor       Processor
	HandlerMap      HandlerMap
	ProcessorDomain string
}

// Server provides HTTP access to an underlying Processor.
type Server struct {
	router          *httprouter.Router
	processorDomain string
}

func (s Server) error(w http.ResponseWriter, err ErrorResponse) {
	w.Header().Set("Content Type", "application/json")
	w.Header().Set("Cache Control", "no store")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (s Server) handle(fn Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("X-OpenGDPR-ProcessorDomain", s.processorDomain)
		err := fn(w, r, p)
		if err != nil {
			if e, ok := err.(ErrorResponse); ok {
				s.error(w, e)
			} else {
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
