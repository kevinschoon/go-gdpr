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
		migrate         = flag.Bool("migrate", false, "migrate the database")
		processorDomain = flag.String("processor_domain", "", "Processor DNS name")
		privateKeyPath  = flag.String("private_key_path", "key.pem", "private cerificate path")
		publicKeyPath   = flag.String("public_key_path", "cert.pem", "public key path")
	)
	flag.Parse()
	db, err := NewDatabase("gdpr.sqlite")
	maybe(err)
	if *migrate {
		log.Println("migrating database")
		log.Println(db.Migrate())
	}
	proc := &Processor{
		db: db,
	}
	signer, err := gdpr.NewSigner(&gdpr.SignerOptions{
		PrivateKeyPath: *privateKeyPath,
	})
	maybe(err)
	verifier, err := gdpr.NewVerifier(&gdpr.VerifierOptions{
		PublicKeyPath: *publicKeyPath,
	})
	maybe(err)
	svr := gdpr.NewServer(&gdpr.ServerOptions{
		ProcessorCertificateBytes: verifier.Key(),
		Signer:          signer,
		ProcessorDomain: *processorDomain,
		Processor:       proc,
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
}
