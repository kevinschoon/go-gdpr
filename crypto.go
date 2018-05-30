package gdpr

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

// Signer accepts a byte array which it
// creates a hash from and generates a
// signature which it base64 encodes as
// a string.
type Signer interface {
	// Generate a signature for the digest
	Sign(body []byte) (string, error)
}

// Verifier accepts a byte array and
// base64 encoded signature. It hashes
// the byte array and compares it's
// decoded value.
type Verifier interface {
	// return the underlying cerificate
	Cert() *x509.Certificate
	// verify the digest of the signature
	Verify(body []byte, signature string) error
}

// KeyOptions specify the path or bytes
// of a public or private key.
type KeyOptions struct {
	KeyPath  string
	KeyBytes []byte
	// Optional byte string to decrypt
	// a private key file.
	Password []byte
}

func MustNewSigner(opts *KeyOptions) Signer {
	signer, err := NewSigner(opts)
	if err != nil {
		panic(err)
	}
	return signer
}

// NewSigner creates a new RSA backed Signer
func NewSigner(opts *KeyOptions) (Signer, error) {
	privateKey := opts.KeyBytes
	if opts.KeyPath != "" {
		raw, err := ioutil.ReadFile(opts.KeyPath)
		if err != nil {
			return nil, err
		}
		privateKey = raw
	}
	block, _ := pem.Decode(privateKey)
	blockBytes := block.Bytes
	// Decode the PEM key if a password is set
	if x509.IsEncryptedPEMBlock(block) {
		b, err := x509.DecryptPEMBlock(block, opts.Password)
		if err != nil {
			return nil, err
		}
		blockBytes = b
	}
	parsed, err := x509.ParsePKCS8PrivateKey(blockBytes)
	if err != nil {
		return nil, err
	}
	privKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unsupported private key")
	}
	return &rsaSigner{privKey: privKey}, nil
}

func MustNewVerifier(opts *KeyOptions) Verifier {
	verifier, err := NewVerifier(opts)
	if err != nil {
		panic(err)
	}
	return verifier
}

// NewVerifier creates a new RSA backed Verifier
func NewVerifier(opts *KeyOptions) (Verifier, error) {
	publicKey := opts.KeyBytes
	if opts.KeyPath != "" {
		raw, err := ioutil.ReadFile(opts.KeyPath)
		if err != nil {
			return nil, err
		}
		publicKey = raw
	}
	block, _ := pem.Decode(publicKey)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	// TODO
	// Currently only support RSA keys
	// Need to consider DSA/ECDSA
	if cert.PublicKeyAlgorithm != x509.RSA {
		return nil, fmt.Errorf("unsupported public key type")
	}
	return &rsaVerifier{publicKey: cert.PublicKey.(*rsa.PublicKey), cert: cert}, nil
}

type rsaSigner struct {
	privKey *rsa.PrivateKey
}

func (s *rsaSigner) Sign(body []byte) (string, error) {
	// hash the digest body
	hashed := sha256.Sum256(body)
	// Create a new PSS signature
	signature, err := rsa.SignPSS(rand.Reader, s.privKey, crypto.SHA256,
		hashed[:], &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto})
	buf := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	_, err = encoder.Write(signature)
	if err != nil {
		return "", err
	}
	err = encoder.Close()
	if err != nil {
		return "", err
	}
	return buf.String(), nil

}

type rsaVerifier struct {
	publicKey *rsa.PublicKey
	cert      *x509.Certificate
}

func (v *rsaVerifier) Verify(body []byte, signature string) error {
	// Decode signature from base64
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	// Hash the body digest
	hashed := sha256.Sum256(body)
	err = rsa.VerifyPSS(v.publicKey, crypto.SHA256, hashed[:],
		decoded, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto})
	if err != nil {
		// Verification failed
		return ErrInvalidRequestSignature(signature, err)
	}
	// Signature is valid
	return nil
}

func (v *rsaVerifier) Cert() *x509.Certificate {
	return v.cert
}

// NoopVerifier is useful to forgo
// all certificate verification.
type NoopVerifier struct{}

func (v NoopVerifier) Verify([]byte, string) error { return nil }

func (v NoopVerifier) Cert() *x509.Certificate { return &x509.Certificate{} }

// NoopSigner is useful to forgo all
// signature generation.
type NoopSigner struct{}

func (s NoopSigner) Sign([]byte) (string, error) { return "", nil }
