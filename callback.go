package gdpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CallbackOptions struct {
	MaxAttempts     int
	Backoff         time.Duration
	ProcessorDomain string
	Signature       string
}

func Callback(cbReq CallbackRequest, opts CallbackOptions) error {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(cbReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", cbReq.StatusCallbackUrl, buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-OpenGDPR-Processor-Domain", opts.ProcessorDomain)
	req.Header.Set("X-OpenGDPR-Signature", opts.Signature)
	// Attempt to make callback
	for i := 0; i < opts.MaxAttempts; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			time.Sleep(opts.Backoff)
			continue
		}
		// Success
		return nil
	}
	return fmt.Errorf("callback timed out for %s", cbReq.StatusCallbackUrl)
}
