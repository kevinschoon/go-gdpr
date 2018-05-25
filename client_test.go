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

var mockError = []byte(`
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
	c := NewClient(&ClientOptions{Endpoint: "http://mock-endpoint"})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockResp)),
	}, nil)
	resp, err := c.Request(&Request{})
	assert.NoError(t, err)
	assert.Equal(t, "example_controller_id", resp.ControllerId)
	assert.Equal(t, "a7551968-d5d6-44b2-9831-815ac9017798", resp.SubjectRequestId)
}

func TestClientError(t *testing.T) {
	c := NewClient(&ClientOptions{Endpoint: "http://mock-endpoint"})
	c.caller = newMockCaller(&http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBuffer(mockError)),
	}, nil)
	_, err := c.Request(&Request{})
	assert.Error(t, err)
	assert.IsType(t, &ErrorResponse{}, err)
}
