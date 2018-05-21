package gdpr

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Gdpr implements the business logic
// for processing GDPR requests.
type Gdpr interface {
	Request(Request) (Response, error)
	Callback(CallbackRequest) error
	Status(string) (StatusResponse, error)
	Cancel(string) (CancellationResponse, error)
}

type Handler func(http.ResponseWriter, *http.Request, httprouter.Params) error

type Builder func(Server, Gdpr) Handler

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
		"/opengdpr_requests": map[string]Builder{
			"GET":    getRequest,
			"POST":   postRequest,
			"DELETE": deleteRequest,
		},
		"/discovery": map[string]Builder{
			"GET": getDiscovery,
		},
	}
}

type ServerOptions struct {
	Identities   []Identity
	SubjectTypes []SubjectType
	Gdpr         Gdpr
	HandlerMap   HandlerMap
}

type Server struct {
	router       *httprouter.Router
	subjectTypes []SubjectType
	identities   []Identity
	handler      Gdpr
}

func (s Server) Error(w http.ResponseWriter, err ErrorResponse) {
	w.Header().Set("Content Type", "application/json")
	w.Header().Set("Cache Control", "no store")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (s Server) handle(fn Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := fn(w, r, p)
		if err != nil {
			s.Error(w, ErrorResponse{Message: err.Error()})
		}
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(opts ServerOptions) Server {
	server := Server{
		router:       httprouter.New(),
		identities:   opts.Identities,
		subjectTypes: opts.SubjectTypes,
	}
	hm := defaultHandlerMap()
	if opts.HandlerMap != nil {
		hm.Merge(opts.HandlerMap)
	}
	for path, methods := range hm {
		for method, builder := range methods {
			server.router.Handle(method, path, server.handle(builder(server, opts.Gdpr)))
		}
	}
	return server
}
