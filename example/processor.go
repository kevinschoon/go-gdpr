package main

import (
	"log"
	"time"

	"github.com/greencase/go-gdpr"
)

// SQLite backed OpenGDPR server implementation
type Processor struct {
	db     *Database
	domain string
}

func (p *Processor) Request(req *gdpr.Request) (*gdpr.Response, error) {
	log.Printf("processing new request %s\n", req.SubjectRequestId)
	dbReq := dbState{
		SubjectRequestId:       req.SubjectRequestId,
		RequestStatus:          string(gdpr.STATUS_PENDING),
		EncodedRequest:         req.Base64(),
		SubmittedTime:          req.SubmittedTime,
		ReceivedTime:           time.Now(),
		StatusCallbackUrls:     req.StatusCallbackUrls,
		ExpectedCompletionTime: time.Now().Add(5 * time.Second),
	}
	return &gdpr.Response{
		ReceivedTime:           dbReq.ReceivedTime,
		SubjectRequestId:       dbReq.SubjectRequestId,
		ExpectedCompletionTime: dbReq.ExpectedCompletionTime,
		EncodedRequest:         dbReq.EncodedRequest,
	}, p.db.Write(dbReq)
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
	req.RequestStatus = string(gdpr.STATUS_CANCELLED)
	err = p.db.Write(*req)
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

func (p *Processor) Process() error {
	pending, err := p.db.Pending()
	if err != nil {
		return err
	}
	log.Printf("processing %d pending requests\n", len(pending))
	doneCh := make(chan struct {
		SubjectRequestId string
		Err              error
	})
	for _, request := range pending {
		// TODO: Process Requests Here!
		go func(request *dbState) {
			cbReq := &gdpr.CallbackRequest{}
			err = gdpr.Callback(cbReq, &gdpr.CallbackOptions{
				MaxAttempts:     3,
				ProcessorDomain: p.domain,
				Backoff:         5 * time.Second,
				Signature:       request.EncodedRequest,
			})
			doneCh <- struct {
				SubjectRequestId string
				Err              error
			}{
				SubjectRequestId: request.SubjectRequestId,
				Err:              err,
			}
		}(request)
	}
	for i := 0; i < len(pending); i++ {
		msg := <-doneCh
		if msg.Err != nil {
			// BUG: The OpenGDPR specification doesn't say what should happen
			// when Callback requests fail so we just mark it as COMPLETED.
			log.Printf("callback for request %s failed: %s\n", msg.SubjectRequestId, msg.Err)
		}
		err = p.db.SetStatus(msg.SubjectRequestId, gdpr.STATUS_COMPLETED)
		if err != nil {
			return err
		}
		log.Printf("request %s marked as completed \n", msg.SubjectRequestId)
	}
	return nil
}
