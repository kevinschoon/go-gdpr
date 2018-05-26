# go-gdpr
[![GoDoc](https://godoc.org/github.com/greencase/go-gdpr?status.svg)](https://godoc.org/github.com/greencase/go-gdpr)

`go-gdpr` is an experimental implementation of the [OpenGDPR](https://www.opengdpr.org) specification for use with the EU [GDPR](https://www.eugdpr.org/) regulation. **Disclaimer: Using this library does not imply accordance with GDPR!** This project is intended to be consumed as a library to aid in the processing of HTTP requests based on the OpenGDPR standard.

## Installation

    go get github.com/greencase/go-gdpr


## Usage

The primary use case for `go-gdpr` is wrapping business logic via the `Processor` interface with the `Server` type. There is an additional `Client` which allows the consumer to access the processor via HTTP. `go-gdpr` may also be useful by providing static typs for OpenGDPR.

### Simple Processor Example

    package main

    import (
        "fmt"
        "github.com/greencase/go-gdpr"
        "net/http"
        "os"
    )

    type Processor struct{}

    func (p *Processor) Request(req *gdpr.Request) (*gdpr.Response, error) {
        // Process the request..
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
        server := gdpr.NewServer(&gdpr.ServerOptions{
            ProcessorDomain: "my-processor-domain.com",
            Processor:       &Processor{},
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
        err := http.ListenAndServe(":4000", server)
        if err != nil {
            fmt.Println("Error: ", err)
            os.Exit(1)
        }
    }


### Stateful Processor Example

The `example` package implements an OpenGDPR processor backed with a SQLite database. It launches a go-gdpr `Server` in a separate Go routine and then polls the database for new requests in the `STATUS_PENDING` state. Note that this technique is not particularly efficient and in a production deployment one would likely prefer a message queue. 

#### Installation

Ensure you have [dep](https://github.com/golang/dep) installed.

    cd example && dep ensure
    go run *.go -migrate

    2018/05/26 16:27:24 migrating database
    2018/05/26 16:27:24 table request already exists
    2018/05/26 16:27:24 server listening @ :4000
    2018/05/26 16:27:24 processing 0 pending requests
    ...

Make a new GDPR request to the processor:

    # Create an erasure request
    curl -X POST -d '{"subject_request_id": "1234", "subject_request_type":"erasure", "subject_identities": [{"identity_type": "email", "identity_format": "raw"}]}' http://localhost:4000/opengdpr_requests

    ...
    {
      "controller_id": "",
      "expected_completion_time": "2018-05-26T16:36:34.075892734+01:00",
      "received_time": "2018-05-26T16:36:29.075892545+01:00",
      "encoded_request": "eyJzdWJqZWN0X3JlcXVlc3RfaWQiOiIxMjM0Iiwic3ViamVjdF9yZXF1ZXN0X3R5cGUiOiJlcmFzdXJlIiwic3VibWl0dGVkX3RpbWUiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImFwaV92ZXJzaW9uIjoiIiwic3RhdHVzX2NhbGxiYWNrX3VybHMiOm51bGwsInN1YmplY3RfaWRlbnRpdGllcyI6W3siaWRlbnRpdHlfdHlwZSI6ImVtYWlsIiwiaWRlbnRpdHlfZm9ybWF0IjoicmF3In1dLCJleHRlbnNpb25zIjpudWxsfQ==",
      "subject_request_id": "1234"
    }

    # Check the status of the request

    curl -v localhost:4000/opengdpr_requests/1234

    ...
            {
                "controller_id":"",
                "expected_completion_time":"2018-05-26T16:36:34.075892734+01:00",
                "subject_request_id":"1234",
                "request_status":"completed",
                "api_version":"0.1",
                "results_url":""
            }

    # Cancel the request

    curl -X DELETE localhost:4000/opengdpr_requests/1234

    ...
        {
        "controller_id":"",
        "subject_request_id":"1234",
        "ReceivedTime":"2018-05-26T16:36:29.075892545+01:00",
        "encoded_request":"eyJzdWJqZWN0X3JlcXVlc3RfaWQiOiIxMjM0Iiwic3ViamVjdF9yZXF1ZXN0X3R5cGUiOiJlcmFzdXJlIiwic3VibWl0dGVkX3RpbWUiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImFwaV92ZXJzaW9uIjoiIiwic3RhdHVzX2NhbGxiYWNrX3VybHMiOm51bGwsInN1YmplY3RfaWRlbnRpdGllcyI6W3siaWRlbnRpdHlfdHlwZSI6ImVtYWlsIiwiaWRlbnRpdHlfZm9ybWF0IjoicmF3In1dLCJleHRlbnNpb25zIjpudWxsfQ==",
        "api_version":"0.1"}


Callback requests are also honored:


    # In a seperate process pane/process:
    nc -l 4001

    # Encode a new request:
    curl -v -X POST -d '{"subject_request_id": "12345", "subject_request_type":"erasure", "subject_identities": [{"identity_type": "email", "identity_format": "raw"}], "status_callback_urls": ["http://localhost:4001"]}' http://localhost:4000/opengdpr_requests
    

    # nc response:

    POST / HTTP/1.1
    Host: localhost:4001
    User-Agent: Go-http-client/1.1
    Content-Length: 178
    X-Opengdpr-Processor-Domain: 
    X-Opengdpr-Signature: eyJzdWJqZWN0X3JlcXVlc3RfaWQiOiJhc2RmMTIzNDdhc2RmYXNkZkZVSUMiLCJzdWJqZWN0X3JlcXVlc3RfdHlwZSI6ImVyYXN1cmUiLCJzdWJtaXR0ZWRfdGltZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwiYXBpX3ZlcnNpb24iOiIiLCJzdGF0dXNfY2FsbGJhY2tfdXJscyI6WyJodHRwOi8vbG9jYWxob3N0OjQwMDEiXSwic3ViamVjdF9pZGVudGl0aWVzIjpbeyJpZGVudGl0eV90eXBlIjoiZW1haWwiLCJpZGVudGl0eV9mb3JtYXQiOiJyYXcifV0sImV4dGVuc2lvbnMiOm51bGx9
    Accept-Encoding: gzip

    {"controller_id":"","expected_completion_time":"0001-01-01T00:00:00Z","status_callback_url":"http://localhost:4001","subject_request_id":"","request_status":"","results_url":""}

## TODO

    * SSL / Certificate Support
    * Certificate Verification
    * Domain/Controller Whitelisting


## Contributing

We are open to any and all contributions so long as they improve the library, feel free to open up a new [issue](https://github.com/greencase/go-gdpr/issues)!



