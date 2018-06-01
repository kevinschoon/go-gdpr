package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/greencase/go-gdpr"
)

func maybe(err error) {
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}

func main() {
	var (
		interval       = flag.String("interval", "1s", "interval to generate new requests at")
		processor      = flag.Bool("processor", false, "run as a processor")
		controller     = flag.Bool("controller", false, "run as a controller")
		migrate        = flag.Bool("migrate", false, "migrate the database")
		privateKeyPath = flag.String("private_key_path", "key.pem", "private cerificate path")
		publicKeyPath  = flag.String("public_key_path", "cert.pem", "public key path")
	)
	flag.Parse()
	db, err := NewDatabase("gdpr.sqlite")
	maybe(err)
	if *migrate {
		log.Println(db.Migrate())
	}
	signer, err := gdpr.NewSigner(&gdpr.KeyOptions{
		KeyPath: *privateKeyPath,
	})
	maybe(err)
	verifier, err := gdpr.NewVerifier(&gdpr.KeyOptions{
		KeyPath: *publicKeyPath,
	})
	maybe(err)
	// Run a Processor
	if *processor {
		proc := &Processor{
			queue:  make(chan *dbState),
			db:     db,
			signer: signer,
		}
		svr := gdpr.NewServer(&gdpr.ServerOptions{
			Signer:    signer,
			Processor: proc,
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
		// Log the signature generated from the response body which
		// is sent to the controller for verification.
		svr.After(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("signature=%s\n", w.Header().Get("X-OpenGDPR-Signature"))
		})
		go func() {
			// Start the HTTP server in the background
			log.Println("server listening @ :4000")
			maybe(http.ListenAndServe(":4000", svr))
		}()
		maybe(proc.Process())
		return
	}
	// Run a Controller
	if *controller {
		sleepInterval, err := time.ParseDuration(*interval)
		maybe(err)
		contr := &Controller{
			client: gdpr.NewClient(&gdpr.ClientOptions{
				Endpoint: "http://localhost:4000",
				Verifier: verifier,
			}),
			responses: map[string]*gdpr.Response{},
		}
		svr := gdpr.NewServer(&gdpr.ServerOptions{
			Verifier:   verifier,
			Controller: contr,
		})
		// Log the signature generated from the processor which is present
		// on each callback to the controller.
		svr.Before(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("signature=%s\n", r.Header.Get("X-OpenGDPR-Signature"))
		})
		// Start the HTTP server in the background
		go func() {
			log.Println("server listening @ :4001")
			maybe(http.ListenAndServe(":4001", svr))
		}()
		for {
			maybe(contr.Request())
			time.Sleep(sleepInterval)
		}
	}
}
