package gdpr

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Processor implements the business logic
// for processing GDPR requests.
type Processor interface {
	Request(*Request) (*Response, error)
	Status(string) (*StatusResponse, error)
	Cancel(string) (*CancellationResponse, error)
}

type Handler func(http.ResponseWriter, *http.Request, httprouter.Params) error

type Builder func(opts ServerOptions) Handler

type HandlerMap map[string]map[string]Builder

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

type ServerOptions struct {
	Identities      []Identity
	SubjectTypes    []SubjectType
	Processor       Processor
	HandlerMap      HandlerMap
	ProcessorDomain string
}

type Server struct {
	router          *httprouter.Router
	subjectTypes    []SubjectType
	identities      []Identity
	handler         Processor
	processorDomain string
}

func (s Server) Error(w http.ResponseWriter, err ErrorResponse) {
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
				s.Error(w, e)
			} else {
				s.Error(w, ErrorResponse{Message: err.Error(), Code: 500})
			}
		}
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.router.ServeHTTP(w, r) }

func NewServer(opts ServerOptions) Server {
	server := Server{
		router:          httprouter.New(),
		identities:      opts.Identities,
		subjectTypes:    opts.SubjectTypes,
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
