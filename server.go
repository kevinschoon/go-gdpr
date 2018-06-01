package gdpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// Handler reads the incoming request body and encodes a
// json payload to resp.
type Handler func(resp io.Writer, req io.Reader, p httprouter.Params) error

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
	handlerFn       http.HandlerFunc
	signer          Signer
	verifier        Verifier
	isProcessor     bool
	isController    bool
	headers         http.Header
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

// respCode maps any successful request with a
// specific status code or returns 200.
func (s *Server) respCode(r *http.Request) int {
	if r.URL.Path == "/opengdpr_requests" && r.Method == "POST" {
		return http.StatusCreated
	}
	return http.StatusOK
}

func (s *Server) error(w http.ResponseWriter, err error) bool {
	if err != nil {
		w.Header().Set("Cache Control", "no-store")
		switch e := err.(type) {
		case ErrorResponse:
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
			return true
		default:
			w.WriteHeader(http.StatusInternalServerError)
			resp := ErrorResponse{Message: e.Error(), Code: 500}
			json.NewEncoder(w).Encode(resp)
			return true
		}
	}
	return false
}

func (s *Server) handle(fn Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.setHeaders(w)
		body := r.Body
		// allocate a new buffer for the response body
		buf := bytes.NewBuffer(nil)
		// If we are serving a controller validate
		// the request before processing and further
		if s.isController {
			// Allocate a new buffer to copy the
			// payload into after verification
			reqBody := bytes.NewBuffer(nil)
			raw, err := ioutil.ReadAll(io.TeeReader(r.Body, reqBody))
			if s.error(w, err) {
				// Failed to decode request body
				return
			}
			if s.error(w, s.verifier.Verify(raw, r.Header.Get("X-OpenGDPR-Signature"))) {
				// Signature verification failed
				return
			}
			// Set the body to the copied original payload
			body = ioutil.NopCloser(reqBody)
		}
		// satisfy the request and process any error
		if s.error(w, fn(buf, body, p)) {
			return
		}
		// If we are serving a processor add a
		// signature of the response payload
		// in our headers.
		if s.isProcessor {
			signature, err := s.signer.Sign(buf.Bytes())
			if err != nil {
				// fatal since we can't sign our own response
				panic(fmt.Sprintf("cannot sign response: %s", err.Error()))
			}
			// Set the response signature
			w.Header().Set("X-OpenGDPR-Signature", signature)
		}
		w.WriteHeader(s.respCode(r))
		// write the response
		w.Write(buf.Bytes())
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.handlerFn(w, r) }

// Before applys any http.HandlerFunc to the request before
// sending it on to the controller/processor. Note that if a
// handler incecepts the body of the request via request.Body
// they must ensure it's contents are added back to request
// or signature verification will fail!
func (s *Server) Before(handlers ...http.HandlerFunc) {
	original := s.handlerFn
	s.handlerFn = func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			handler(w, r)
		}
		original(w, r)
	}
}

// After applys any http.HandlerFunc to the request after
// it has been handled by the controller/processor.
func (s *Server) After(handlers ...http.HandlerFunc) {
	original := s.handlerFn
	s.handlerFn = func(w http.ResponseWriter, r *http.Request) {
		original(w, r)
		for _, handler := range handlers {
			handler(w, r)
		}
	}
}

// NewServer returns a server type that statisfies the
// http.Handler interface.
func NewServer(opts *ServerOptions) *Server {
	server := &Server{
		signer:          opts.Signer,
		verifier:        opts.Verifier,
		isProcessor:     hasProcessor(opts),
		isController:    hasController(opts),
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
	server.handlerFn = router.ServeHTTP
	return server
}

func hasController(opts *ServerOptions) bool {
	return opts.Controller != nil
}

func hasProcessor(opts *ServerOptions) bool {
	return opts.Processor != nil
}
