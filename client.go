package gdpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type caller interface {
	Call(string, string, io.Reader) (*http.Response, error)
}

type defaultCaller struct {
	client  *http.Client
	headers map[string]string
}

func (d defaultCaller) Call(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for key, value := range d.headers {
		req.Header.Set(key, value)
	}
	return d.client.Do(req)
}

// ClientOptions conifigure a Client.
type ClientOptions struct {
	Endpoint string
	Client   *http.Client
}

// Client is an HTTP helper client for making requests
// to an OpenGDPR processor server.
type Client struct {
	endpoint string
	caller   caller
}

func (c *Client) json(resp *http.Response, err error, v interface{}) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		// Try to decode an internal ErrorResponse type
		// then fall back to returning the body as a string.
		errResp := &ErrorResponse{Code: resp.StatusCode}
		err = json.Unmarshal(raw, errResp)
		if err == nil {
			return errResp
		}
		return fmt.Errorf("Server responded with %d: %s", resp.StatusCode, string(raw))
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(raw, v)
}

// Request makes a performs a new GDPR request.
func (c *Client) Request(req *Request) (*Response, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return nil, err
	}
	reqResp := &Response{}
	resp, err := c.caller.Call("POST", c.endpoint+"/opengdpr_requests", buf)
	return reqResp, c.json(resp, err, reqResp)
}

// Status checks the status of an existing GDPR request.
func (c *Client) Status(id string) (*StatusResponse, error) {
	statResp := &StatusResponse{}
	resp, err := c.caller.Call("GET", c.endpoint+"/opengdpr_requests/"+id, nil)
	return statResp, c.json(resp, err, statResp)
}

// Cancel cancels an existing GDPR request.
func (c *Client) Cancel(id string) (*CancellationResponse, error) {
	cancelResp := &CancellationResponse{}
	resp, err := c.caller.Call("DELETE", c.endpoint+"/opengdpr_requests/"+id, nil)
	return cancelResp, c.json(resp, err, cancelResp)
}

// Discovery describes the remote OpenGDPR speciication.
func (c *Client) Discovery() (*DiscoveryResponse, error) {
	discResp := &DiscoveryResponse{}
	resp, err := c.caller.Call("GET", c.endpoint+"/discovery", nil)
	return discResp, c.json(resp, err, discResp)
}

// NewClient returns a new OpenGDPR client.
func NewClient(opts *ClientOptions) *Client {
	cli := opts.Client
	if cli == nil {
		cli = http.DefaultClient
	}
	client := &Client{
		caller: &defaultCaller{
			client: cli,
			headers: map[string]string{
				"GDPR Version": ApiVersion,
				"Content-Type": "Application/JSON",
			},
		},
		endpoint: opts.Endpoint,
	}
	return client
}
