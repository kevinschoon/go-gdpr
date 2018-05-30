# Examples

This directory contains a few more detailed examples of usage.

## Simple Request / Callback

Make a single request and wait for an issued callback.

```bash
cd ./examples
# Launch a Processor server waiting for a Request
go run processor.go &
# Generate a GDPR request
go run controller.go
# >> PROCESSOR: got request: request-1234! 
# >> CONTROLLER: got callback: request-1234:completed!
```

## Stateful

Launch a SQLite backed stateful processor; ensure you have [dep](https://github.com/golang/dep) installed.

### Run

```bash
cd ./examples/stateful
dep ensure
go run *.go -help
# Generate self-signed RSA keys
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
# <ignore all fields> (enter..enter..)
# Launch the Processor
go run *.go -processor
# Open a new pane/terminal window
go run *.go -controller -interval 100ms # Generate 10 req/sec
# >> 2018/05/30 20:32:47 processing new request d6b5caf7-f170-4e06-93aa-2cbbb863bd09
# >> 2018/05/30 20:32:47 request dd1bb099-7df8-4a4b-b434-f01c459b3709 marked as completed 
# >> 2018/05/30 20:32:47 sending callback: http://localhost:4001/opengdpr_callbacks
```
