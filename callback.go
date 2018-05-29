package gdpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CallBackOptions configure an HTTP callback
type CallbackOptions struct {
	MaxAttempts     int
	Backoff         time.Duration
	ProcessorDomain string
	Client          *http.Client
	Signer          Signer
}

// Callback sends the CallbackRequest type to the configured
// StatusCallbackUrl. If it fails to deliver in n attempts or
// the request is invalid it will return an error.
func Callback(cbReq *CallbackRequest, opts *CallbackOptions) error {
	client := opts.Client
	if client == nil {
		client = http.DefaultClient
	}
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(cbReq)
	if err != nil {
		return err
	}
	signature, err := opts.Signer.Sign(buf.Bytes())
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", cbReq.StatusCallbackUrl, buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-OpenGDPR-Processor-Domain", opts.ProcessorDomain)
	req.Header.Set("X-OpenGDPR-Signature", signature)
	// Attempt to make callback
	for i := 0; i < opts.MaxAttempts; i++ {
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			time.Sleep(opts.Backoff)
			continue
		}
		// Success
		return nil
	}
	return fmt.Errorf("callback timed out for %s", cbReq.StatusCallbackUrl)
}
