package gdpr

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockResp = []byte(`
{
    "controller_id":"example_controller_id",
    "expected_completion_time":"2018-11-01T15:00:01Z",
    "received_time":"2018-10-02T15:00:01Z",
    "encoded_request":"<BASE64 ENCODED REQUEST>",
    "subject_request_id":"a7551968-d5d6-44b2-9831-815ac9017798"
}
`)

var mockCancellationResp = []byte(`
{
  "controller_id": "example_controller_id",
  "subject_request_id": "a7551968-d5d6-44b2-9831-815ac9017798",
  "received_time": "2018-10-02T15:00:01Z",
  "api_version": "1.0"
}
`)

var mockStatusResp = []byte(`
{
    "controller_id":"example_controller_id",
    "expected_completion_time":"2018-11-01T15:00:01Z",
    "subject_request_id":"a7551968-d5d6-44b2-9831-815ac9017798",
    "request_status":"pending",
    "api_version":"1.0",
    "results_url":"https://exampleprocessor.com/secure/d188d4ba-12db-48a0-898c-cd0f8ba7b345"
}
`)

var mockDiscoveryResp = []byte(`
{
   "api_version":"1.0",
   "supported_identities":[
      {
         "identity_type":"email",
         "identity_format":"raw"
      },
      {
         "identity_type":"email",
         "identity_format":"sha256"
      }
   ],
   "supported_subject_request_types":[
      "erasure"
   ],
   "processor_certificate":"https://exampleprocessor.com/cert.pem"
}
`)

var mockErrorResp = []byte(`
{
  "error": {
    "code": 400,
    "message": "subject_request_id field is required",
    "errors": [
      {
        "domain": "Validation",
        "reason": "IllegalArgumentException",
        "message": "subject_request_id field is required."
      }
    ]
  }
}
`)

type mockCaller struct {
	resp *http.Response
	err  error
}

func (m mockCaller) Call(string, string, io.Reader) (*http.Response, error) {
	return m.resp, m.err
}

func newMockCaller(resp *http.Response, err error) *mockCaller {
	return &mockCaller{resp: resp, err: err}
}

func TestClientRequest(t *testing.T) {
	c := NewClient(&ClientOptions{Verifier: NoopVerifier{}})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockResp)),
	}, nil)
	resp, err := c.Request(&Request{})
	assert.NoError(t, err)
	assert.Equal(t, "example_controller_id", resp.ControllerId)
	assert.Equal(t, "a7551968-d5d6-44b2-9831-815ac9017798", resp.SubjectRequestId)
}

func TestClientStatus(t *testing.T) {
	c := NewClient(&ClientOptions{Verifier: NoopVerifier{}})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockStatusResp)),
	}, nil)
	resp, err := c.Status("1234")
	assert.NoError(t, err)
	assert.Equal(t, "a7551968-d5d6-44b2-9831-815ac9017798", resp.SubjectRequestId)
	assert.Equal(t, "example_controller_id", resp.ControllerId)
	assert.Equal(t, "1.0", resp.ApiVersion)
	assert.Equal(t, STATUS_PENDING, resp.RequestStatus)
}

func TestClientCancel(t *testing.T) {
	c := NewClient(&ClientOptions{Verifier: NoopVerifier{}})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockCancellationResp)),
	}, nil)
	resp, err := c.Cancel("1234")
	assert.NoError(t, err)
	assert.Equal(t, "a7551968-d5d6-44b2-9831-815ac9017798", resp.SubjectRequestId)
	assert.Equal(t, "example_controller_id", resp.ControllerId)
	assert.Equal(t, "1.0", resp.ApiVersion)
}

func TestClientDiscover(t *testing.T) {
	c := NewClient(&ClientOptions{Verifier: NoopVerifier{}})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockDiscoveryResp)),
	}, nil)
	resp, err := c.Discovery()
	assert.NoError(t, err)
	assert.Equal(t, IDENTITY_EMAIL, resp.SupportedIdentities[0].Type)
	assert.Equal(t, FORMAT_RAW, resp.SupportedIdentities[0].Format)
	assert.Equal(t, IDENTITY_EMAIL, resp.SupportedIdentities[1].Type)
	assert.Equal(t, FORMAT_SHA256, resp.SupportedIdentities[1].Format)
	assert.Equal(t, SUBJECT_ERASURE, resp.SupportedSubjectRequestTypes[0])
}

func TestClientError(t *testing.T) {
	c := NewClient(&ClientOptions{Verifier: NoopVerifier{}})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 500,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockErrorResp)),
	}, nil)
	_, err := c.Request(&Request{})
	assert.Error(t, err)
	assert.IsType(t, &ErrorResponse{}, err)
	assert.Equal(t, "IllegalArgumentException", err.(*ErrorResponse).Errors[0].Reason)
	assert.Equal(t, "subject_request_id field is required.", err.(*ErrorResponse).Errors[0].Message)
}
