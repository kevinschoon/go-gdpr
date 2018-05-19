package main

import (
	"github.com/greencase/go-gdpr"
)

func runClient() {
	client := gdpr.NewClient(gdpr.ClientOptions{Endpoint: "127.0.0.1:4000"})
	client.Request(gdpr.Request{
		SubjectRequestId: "abdc-1234",
	}.)
}
