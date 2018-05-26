package gdpr

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockResponseBody = []byte(`
{
  "subject_request_id": "a7551968-d5d6-44b2-9831-815ac9017798",
  "subject_request_type": "erasure",
  "submitted_time": "2018-10-02T15:00:00Z",
  "subject_identities": [
    {
      "identity_type": "email",
      "identity_value": "johndoe@example.com",
      "identity_format": "raw"
    }
  ],
  "api_version": "1.0",
  "status_callback_urls": [
    "https://examplecontroller.com/opengdpr_callbacks"
  ],
  "extensions": {
    "example-processor.com": {
      "foo-processor-custom-id":123456,
      "property_id": "123456"
    },
    "example-other-processor.com": {
      "foo-other-processor-custom-id":654321
    }
  }
}
`)

type mockProcessor struct {
	response             *Response
	statusResponse       *StatusResponse
	cancellationResponse *CancellationResponse
	err                  error
}

func (m mockProcessor) Request(req *Request) (*Response, error)   { return m.response, m.err }
func (m mockProcessor) Status(id string) (*StatusResponse, error) { return m.statusResponse, m.err }
func (m mockProcessor) Cancel(id string) (*CancellationResponse, error) {
	return m.cancellationResponse, m.err
}

func newServer() (*Server, *mockProcessor) {
	proc := &mockProcessor{
		response: &Response{
			SubjectRequestId: "1234",
			ControllerId:     "c-1234",
			EncodedRequest:   "1234",
		},
		statusResponse: &StatusResponse{
			SubjectRequestId: "1234",
			ControllerId:     "c-1234",
			RequestStatus:    STATUS_PENDING,
			ApiVersion:       ApiVersion,
		},
		cancellationResponse: &CancellationResponse{
			SubjectRequestId: "1234",
			ControllerId:     "c-1234",
			EncodedRequest:   "1234",
			ApiVersion:       ApiVersion,
		},
	}
	return NewServer(&ServerOptions{
		Processor: proc,
		SubjectTypes: []SubjectType{
			SUBJECT_ERASURE,
		},
		Identities: []Identity{
			Identity{
				Type:   IDENTITY_EMAIL,
				Format: FORMAT_RAW,
			},
		},
	}), proc
}

func TestServerDiscovery(t *testing.T) {
	server, _ := newServer()
	r := httptest.NewRequest("GET", "/discovery", bytes.NewBuffer(mockResponseBody))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	t.Log(w.Body.String())
	assert.Equal(t, 200, w.Code)
	resp := &DiscoveryResponse{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, SUBJECT_ERASURE, resp.SupportedSubjectRequestTypes[0])
	assert.Equal(t, IDENTITY_EMAIL, resp.SupportedIdentities[0].Type)
	assert.Equal(t, FORMAT_RAW, resp.SupportedIdentities[0].Format)
	assert.Equal(t, ApiVersion, resp.ApiVersion)
}

func TestServerRequest(t *testing.T) {
	server, _ := newServer()
	r := httptest.NewRequest("POST", "/opengdpr_requests", bytes.NewBuffer(mockResponseBody))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	t.Log(w.Body.String())
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "application/json", w.Header().Get("Accept"))
	resp := &Response{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
}

func TestServerStatus(t *testing.T) {
	server, _ := newServer()
	r := httptest.NewRequest("GET", "/opengdpr_requests/1234", bytes.NewBuffer(mockResponseBody))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	t.Log(w.Body.String())
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "application/json", w.Header().Get("Accept"))
	resp := &StatusResponse{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
}

func TestServerCancel(t *testing.T) {
	server, _ := newServer()
	r := httptest.NewRequest("GET", "/opengdpr_requests/1234", bytes.NewBuffer(mockResponseBody))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	t.Log(w.Body.String())
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "application/json", w.Header().Get("Accept"))
	resp := &CancellationResponse{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
}
