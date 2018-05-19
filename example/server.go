package main

import (
	"github.com/greencase/go-gdpr"
	//"github.com/satori/go.uuid"
	"log"
	"net/http"
	"time"
)

type Handler struct {
}

func (h Handler) Request(req gdpr.Request) (gdpr.Response, error) {
	log.Println("processing request")
	return gdpr.Response{
		ExpectedCompletionTime: time.Now().Add(10 * time.Second),
		ReceivedTime:           time.Now(),
		SubjectRequestId:       req.SubjectRequestId,
		EncodedRequest:         req.Base64(),
	}, nil
}

func (h Handler) Callback(req gdpr.CallbackRequest) error {
	// TODO: Implement Callback Functionality
	return nil
}

func (h Handler) Status(string) (gdpr.StatusResponse, error) {
	return gdpr.StatusResponse{}, nil
}

func (h Handler) Cancel(string) (gdpr.CancellationResponse, error) {
	return gdpr.CancellationResponse{}, nil
}

func runServer() {
	server := gdpr.NewServer(gdpr.ServerOptions{
		Handler: Handler{},
	})
	log.Println("server listening @ :4000")
	maybe(http.ListenAndServe(":4000", server))
}
