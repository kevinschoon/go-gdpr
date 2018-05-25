package gdpr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ClientOptions struct {
	Endpoint string
}

type Client struct {
	cli      *http.Client
	endpoint string
	headers  map[string]string
}

func (c Client) req(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}

func (c Client) json(r *http.Request, v interface{}) error {
	r.Header.Add("Content-Type", "Application/JSON")
	resp, err := c.cli.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		errResp := ErrorResponse{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return errResp
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c Client) RoundTrip(r *http.Request) (*http.Response, error) {
	for key, value := range c.headers {
		r.Header.Add(key, value)
	}
	return http.DefaultTransport.RoundTrip(r)
}

func (c Client) Request(req Request) (Response, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return Response{}, err
	}
	resp := Response{}
	return resp, c.json(c.req("POST", c.endpoint+"/opengdpr_requests", buf), &resp)
}

func (c Client) Status(id string) (StatusResponse, error) {
	resp := StatusResponse{}
	return resp, c.json(c.req("GET", c.endpoint+"/opengdpr_requests/"+id, nil), &resp)
}

func (c Client) Cancel(id string) (CancellationResponse, error) {
	resp := CancellationResponse{}
	return resp, c.json(c.req("DELETE", c.endpoint+"/opengdpr_requests/"+id, nil), &resp)
}

func (c Client) Discovery() (DiscoveryResponse, error) {
	resp := DiscoveryResponse{}
	return resp, c.json(c.req("GET", c.endpoint+"/discovery", nil), &resp)
}

func NewClient(opts ClientOptions) Client {
	return Client{
		cli: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint: opts.Endpoint,
		headers: map[string]string{
			"GDPR Version": ApiVersion,
		},
	}
}
