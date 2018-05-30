package main

import (
	"fmt"
	"net/http"

	"github.com/greencase/go-gdpr"
)

type Processor struct {
	shutdown chan bool
}

func (p *Processor) Request(req *gdpr.Request) (*gdpr.Response, error) {
	fmt.Printf("PROCESSOR: got request: %s! \n", req.SubjectRequestId)
	// Process the request..
	err := gdpr.Callback(&gdpr.CallbackRequest{
		SubjectRequestId:  req.SubjectRequestId,
		StatusCallbackUrl: req.StatusCallbackUrls[0],
		RequestStatus:     gdpr.STATUS_COMPLETED,
	}, &gdpr.CallbackOptions{
		MaxAttempts: 1,
		Signer:      gdpr.NoopSigner{},
	})
	if err != nil {
		panic(fmt.Sprintf("callback failed: %s", err))
	}
	go func() {
		p.shutdown <- true
	}()
	return nil, nil
}

func (p *Processor) Status(id string) (*gdpr.StatusResponse, error) {
	// Check the status of the request..
	return nil, nil
}

func (p *Processor) Cancel(id string) (*gdpr.CancellationResponse, error) {
	// Cancel the request..
	return nil, nil
}

func main() {
	processor := &Processor{shutdown: make(chan bool)}
	server := gdpr.NewServer(&gdpr.ServerOptions{
		Signer:          gdpr.NoopSigner{},
		ProcessorDomain: "my-processor-domain.com",
		Processor:       processor,
		Identities: []gdpr.Identity{
			gdpr.Identity{
				Type:   gdpr.IDENTITY_EMAIL,
				Format: gdpr.FORMAT_RAW,
			},
		},
		SubjectTypes: []gdpr.SubjectType{
			gdpr.SUBJECT_ACCESS,
			gdpr.SUBJECT_ERASURE,
			gdpr.SUBJECT_PORTABILITY,
		},
	})
	go http.ListenAndServe(":4000", server)
	<-processor.shutdown
}
