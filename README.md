# go-gdpr
[![CircleCI](https://circleci.com/gh/greencase/go-gdpr/tree/master.svg?style=svg)](https://circleci.com/gh/greencase/go-gdpr/tree/master)
[![GoDoc](https://godoc.org/github.com/greencase/go-gdpr?status.svg)](https://godoc.org/github.com/greencase/go-gdpr)

`go-gdpr` is an implementation of the [OpenGDPR](https://www.opengdpr.org) specification for use with the EU [GDPR](https://www.eugdpr.org/) regulation. 

**Disclaimer: Using this library does not imply accordance with GDPR!**

## Installation


    go get github.com/greencase/go-gdpr


## Usage

`go-gdpr` is intended to be used as a library by exposing business logic wrapped as middleware through an HTTP interface that meets the specifications in the OpenGDPR standard. There is no single concise way to achieve GDPR "compliance" per-say without implementing platform specific processes; `go-gdpr` only provides convenient components for fulfilling requests deemed mandatory by the GDPR legislation.


### Concepts

The two major concepts used in this library are the  `Controller` and `Processor` types. Their definitions are listed below and borrowed directly from the OpenGDPR specification.

#### Controller

> An entity which makes the decision about what personal data  will be processed and the types of processing that will be done with respect to that personal data. The Data Controller receives Data Subject requests from the Data Subjects and validates them. 

#### Processor

> The organization that processes data pursuant to the instructions of the Controller on behalf of the Controller. The Data Processor receives data subject requests via RESTful endpoints and is responsible for fulfilling requests.

## Usage

The primary use case for `go-gdpr` is wrapping business logic via the `Processor` and `Controller` interface with the `Server` type. There is an additional `Client` which allows the consumer to access the processor via HTTP calls. This library might also be useful by providing static typing for other server implementations. See the [example](https://github.com/greencase/go-gdpr/tree/master/example) section for a more thorough introduction.

### Simple Processor Example

A basic `Processor` can be implemented with just three methods:

```go
package main

import (
	"net/http"

	"github.com/greencase/go-gdpr"
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
		Signer:          gdpr.NoopSigner{},
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
	http.ListenAndServe(":4000", server)
}
```

### Simple Controller Example

A `Controller` only requires a single method, although a `Client` must also be used to communicate with the `Processor`:

```go
package main

import (
	"net/http"

	"github.com/greencase/go-gdpr"
)

type Controller struct{}

func (c *Controller) Callback(req *gdpr.CallbackRequest) error {
	// Process the callback..
	return nil
}

func main() {
	server := gdpr.NewServer(&gdpr.ServerOptions{
		Controller: &Controller{},
		Verifier:   gdpr.NoopVerifier{},
	})
	http.ListenAndServe(":4001", server)
}
```

## Contributing

We are open to any and all contributions so long as they improve the library, feel free to open up a new [issue](https://github.com/greencase/go-gdpr/issues)!



