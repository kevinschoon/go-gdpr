package main

import (
	"log"
	"sync"

	"github.com/satori/go.uuid"

	"github.com/greencase/go-gdpr"
)

type Controller struct {
	mu        sync.RWMutex
	client    *gdpr.Client
	responses map[string]*gdpr.Response
}

func (c *Controller) Callback(cb *gdpr.CallbackRequest) error {
	log.Printf("received callback request for %s: %s\n", cb.SubjectRequestId, cb.RequestStatus)
	if cb.RequestStatus == gdpr.STATUS_COMPLETED {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.responses, cb.SubjectRequestId)
	}
	return nil
}

// Request generates a random/fake GDPR request
func (c *Controller) Request() error {
	req := &gdpr.Request{
		ApiVersion:       gdpr.ApiVersion,
		SubjectRequestId: uuid.NewV4().String(),
		SubjectIdentities: []gdpr.Identity{
			gdpr.Identity{
				Type:   gdpr.IDENTITY_EMAIL,
				Format: gdpr.FORMAT_RAW,
				Value:  "username@some-email.com",
			},
		},
		SubjectRequestType: gdpr.SUBJECT_ERASURE,
		StatusCallbackUrls: []string{
			"http://localhost:4001/opengdpr_callbacks",
		},
	}
	resp, err := c.client.Request(req)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.responses[resp.SubjectRequestId] = resp
	log.Printf("sent new gdpr request: %s", resp.SubjectRequestId)
	return nil
}
