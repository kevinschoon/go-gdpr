package main

import (
	"log"
	"time"

	"github.com/greencase/go-gdpr"
)

// Done is used to communicate a callback
// was completed.
type Done struct {
	SubjectRequestId string
	Err              error
}

// SQLite backed OpenGDPR processor implementation
type Processor struct {
	db     *Database
	domain string
	queue  chan *dbState
	signer gdpr.Signer
}

func (p *Processor) Request(req *gdpr.Request) (*gdpr.Response, error) {
	log.Printf("processing new request %s\n", req.SubjectRequestId)
	dbReq := &dbState{
		SubjectRequestId:       req.SubjectRequestId,
		RequestStatus:          string(gdpr.STATUS_PENDING),
		EncodedRequest:         req.Base64(),
		SubmittedTime:          req.SubmittedTime,
		ReceivedTime:           time.Now(),
		StatusCallbackUrls:     req.StatusCallbackUrls,
		ExpectedCompletionTime: time.Now().Add(5 * time.Second),
	}
	err := p.db.Write(dbReq)
	if err != nil {
		return nil, err
	}
	p.queue <- dbReq
	return &gdpr.Response{
		ReceivedTime:           dbReq.ReceivedTime,
		SubjectRequestId:       dbReq.SubjectRequestId,
		ExpectedCompletionTime: dbReq.ExpectedCompletionTime,
		EncodedRequest:         dbReq.EncodedRequest,
	}, nil
}

func (p *Processor) Status(requestId string) (*gdpr.StatusResponse, error) {
	log.Printf("processing status request %s\n", requestId)
	req, err := p.db.Read(requestId)
	if err != nil {
		return nil, err
	}
	return &gdpr.StatusResponse{
		SubjectRequestId:       req.SubjectRequestId,
		RequestStatus:          gdpr.RequestStatus(req.RequestStatus),
		ExpectedCompletionTime: req.ExpectedCompletionTime,
		ApiVersion:             gdpr.ApiVersion,
	}, nil
}

func (p *Processor) Cancel(requestId string) (*gdpr.CancellationResponse, error) {
	log.Printf("processing cancellation %s\n", requestId)
	req, err := p.db.Read(requestId)
	if err != nil {
		return nil, err
	}
	err = p.db.SetStatus(req.SubjectRequestId, gdpr.STATUS_CANCELLED)
	if err != nil {
		return nil, err
	}
	return &gdpr.CancellationResponse{
		SubjectRequestId: requestId,
		ApiVersion:       gdpr.ApiVersion,
		ReceivedTime:     req.ReceivedTime,
		EncodedRequest:   req.EncodedRequest,
	}, nil
}

func (p *Processor) process(request *dbState, doneCh chan Done) {
	for _, cbUrl := range request.StatusCallbackUrls {
		log.Printf("sending callback: %s", cbUrl)
		cbReq := &gdpr.CallbackRequest{
			SubjectRequestId:  request.SubjectRequestId,
			RequestStatus:     gdpr.STATUS_COMPLETED,
			StatusCallbackUrl: cbUrl,
		}
		err := gdpr.Callback(cbReq, &gdpr.CallbackOptions{
			MaxAttempts:     3,
			ProcessorDomain: p.domain,
			Backoff:         5 * time.Second,
			Signer:          p.signer,
		})
		doneCh <- Done{
			SubjectRequestId: request.SubjectRequestId,
			Err:              err,
		}
	}
}

func (p *Processor) Process() error {
	doneCh := make(chan Done)
	for {
		select {
		case req := <-p.queue:
			go p.process(req, doneCh)
		case done := <-doneCh:
			if done.Err != nil {
				// BUG: The OpenGDPR specification doesn't say what should happen
				// when Callback requests fail so we just mark it as COMPLETED.
				log.Printf("callback for request %s failed: %s\n", done.SubjectRequestId, done.Err)
			}
			err := p.db.SetStatus(done.SubjectRequestId, gdpr.STATUS_COMPLETED)
			if err != nil {
				return err
			}
			log.Printf("request %s marked as completed \n", done.SubjectRequestId)
		}
	}
	return nil
}
