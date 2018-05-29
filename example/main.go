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
	signer, err := gdpr.NewSigner(&gdpr.SignerOptions{
		PrivateKeyPath: *privateKeyPath,
	})
	maybe(err)
	verifier, err := gdpr.NewVerifier(&gdpr.VerifierOptions{
		PublicKeyPath: *publicKeyPath,
	})
	maybe(err)
	if *processor {
		proc := &Processor{db: db, signer: signer}
		svr := gdpr.NewServer(&gdpr.ServerOptions{
			ProcessorCertificateBytes: verifier.Key(),
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
		go func() {
			// Start the HTTP server in the background
			log.Println("server listening @ :4000")
			maybe(http.ListenAndServe(":4000", svr))
		}()
		for {
			maybe(proc.Process())
			time.Sleep(5 * time.Second)
		}
		return
	}
	if *controller {
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
		// Start the HTTP server in the background
		go func() {
			log.Println("server listening @ :4001")
			maybe(http.ListenAndServe(":4001", svr))
		}()
		for {
			maybe(contr.Request())
			time.Sleep(2 * time.Second)
		}
	}
}
