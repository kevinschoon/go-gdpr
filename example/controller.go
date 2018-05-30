package main

import (
	"fmt"
	"net/http"

	"github.com/greencase/go-gdpr"
)

type Controller struct {
	shutdown chan bool
}

func (c *Controller) Callback(req *gdpr.CallbackRequest) error {
	// Process the callback..
	fmt.Printf("CONTROLLER: got callback: %s:%s!\n", req.SubjectRequestId, req.RequestStatus)
	go func() {
		c.shutdown <- true
	}()
	return nil
}

func main() {
	controller := &Controller{shutdown: make(chan bool)}
	server := gdpr.NewServer(&gdpr.ServerOptions{
		Controller: controller,
		Verifier:   gdpr.NoopVerifier{},
	})
	go http.ListenAndServe(":4001", server)
	client := gdpr.NewClient(&gdpr.ClientOptions{
		Endpoint: "http://localhost:4000",
		Verifier: gdpr.NoopVerifier{},
	})
	_, err := client.Request(&gdpr.Request{
		SubjectRequestId:   "request-1234",
		SubjectRequestType: gdpr.SUBJECT_ACCESS,
		SubjectIdentities: []gdpr.Identity{
			gdpr.Identity{
				Type:   gdpr.IDENTITY_EMAIL,
				Format: gdpr.FORMAT_RAW,
				Value:  "user@provider.com",
			},
		},
		StatusCallbackUrls: []string{
			"http://localhost:4001/opengdpr_callbacks",
		},
	})
	if err != nil {
		panic(fmt.Sprintf("request failed: %s", err))
	}
	<-controller.shutdown
}
