package gdpr

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler implements the business logic
// for processing GDPR requests.
type Handler interface {
	Request(Request) (Response, error)
	Callback(CallbackRequest) error
	Status(string) (StatusResponse, error)
	Cancel(string) (CancellationResponse, error)
}

type ServerOptions struct {
	Identities   []Identity
	SubjectTypes []SubjectType
	Handler      Handler
}

type Server struct {
	subjectTypes []SubjectType
	identities   []Identity
	handler      Handler
}

func (s Server) RequestGET(w http.ResponseWriter, r *http.Request) error {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 2 {
		return ErrorResponse{Code: 400, Message: "no id specified"}
	}
	resp, err := s.handler.Status(path[1])
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(resp)
}

func (s Server) RequestPOST(w http.ResponseWriter, r *http.Request) error {
	req := &Request{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return err
	}
	resp, err := s.handler.Request(*req)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(resp)
}

func (s Server) RequestDELETE(w http.ResponseWriter, r *http.Request) error {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 2 {
		return ErrorResponse{Code: 400, Message: "no id specified"}
	}
	resp, err := s.handler.Cancel(path[1])
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(resp)
}

func (s Server) DiscoveryGET(w http.ResponseWriter, r *http.Request) error {
	resp := DiscoveryResponse{
		ApiVersion:                   ApiVersion,
		SupportedSubjectRequestTypes: s.subjectTypes,
		SupportedIdentities:          s.identities,
	}
	return json.NewEncoder(w).Encode(resp)
}

func (s Server) CallbackPOST(w http.ResponseWriter, r *http.Request) error {
	req := CallbackRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}
	return s.handler.Callback(req)
}

func (s Server) Error(w http.ResponseWriter, err ErrorResponse) {
	w.Header().Set("Content Type", "application/json")
	w.Header().Set("Cache Control", "no store")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) == 0 {
		s.Error(w, ErrorResponse{Code: 404})
		return
	}
	var err error
	switch path[0] {
	case "/opengdpr_requests":
		switch r.Method {
		case "GET":
			err = s.RequestGET(w, r)
		case "POST":
			err = s.RequestPOST(w, r)
		case "DELETE":
			err = s.RequestDELETE(w, r)
		default:
			err = ErrorResponse{Code: 400, Message: "unsupported method"}
		}
	case "/discovery":
		switch r.Method {
		case "GET":
			err = s.DiscoveryGET(w, r)
		}
	default:
		err = ErrorResponse{Code: 404, Message: "unknown request"}
	}
	if err != nil {
		switch e := err.(type) {
		case ErrorResponse:
			s.Error(w, e)
		default:
			s.Error(w, ErrorResponse{Code: 500, Message: e.Error()})
		}
	}
}

func NewServer(opts ServerOptions) Server {
	return Server{
		handler:      opts.Handler,
		identities:   opts.Identities,
		subjectTypes: opts.SubjectTypes,
	}
}
