package gdpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// ClientOptions conifigure a Client.
type ClientOptions struct {
	Endpoint string
	Client   *http.Client
}

// Client is an HTTP helper client for making requests
// to an OpenGDPR processor server.
type Client struct {
	client   *http.Client
	endpoint string
	headers  map[string]string
}

func (c *Client) req(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	return req
}

func (c *Client) json(r *http.Request, v interface{}) error {
	r.Header.Add("Content-Type", "Application/JSON")
	resp, err := c.client.Do(r)
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
	resp := &Response{}
	return resp, c.json(c.req("POST", c.endpoint+"/opengdpr_requests", buf), resp)
}

// Status checks the status of an existing GDPR request.
func (c *Client) Status(id string) (*StatusResponse, error) {
	resp := &StatusResponse{}
	return resp, c.json(c.req("GET", c.endpoint+"/opengdpr_requests/"+id, nil), resp)
}

// Cancel cancels an existing GDPR request.
func (c *Client) Cancel(id string) (*CancellationResponse, error) {
	resp := &CancellationResponse{}
	return resp, c.json(c.req("DELETE", c.endpoint+"/opengdpr_requests/"+id, nil), resp)
}

// Discovery describes the remote OpenGDPR speciication.
func (c *Client) Discovery() (*DiscoveryResponse, error) {
	resp := &DiscoveryResponse{}
	return resp, c.json(c.req("GET", c.endpoint+"/discovery", nil), resp)
}

// NewClient returns a new OpenGDPR client.
func NewClient(opts *ClientOptions) *Client {
	client := &Client{
		client:   opts.Client,
		endpoint: opts.Endpoint,
		headers: map[string]string{
			"GDPR Version": ApiVersion,
		},
	}
	if client.client == nil {
		client.client = http.DefaultClient
	}
	return client
}
