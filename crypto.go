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

// Signer generates a cryptographic signature
// of a digest of bytes.
type Signer interface {
	// Generate a signature for the digest
	Sign(body []byte) (string, error)
}

// Verifier verifies the signature of a digest
// of bytes.
type Verifier interface {
	// return the underlying public key
	Key() []byte
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
	if cert.PublicKeyAlgorithm != x509.RSA {
		return nil, fmt.Errorf("unsupported public key type")
	}
	return &rsaVerifier{publicKey: cert.PublicKey.(*rsa.PublicKey), pemBlock: block}, nil
}

type rsaSigner struct {
	privKey *rsa.PrivateKey
}

func (s *rsaSigner) Sign(body []byte) (string, error) {
	// Create a hash of the base64
	// encoded JSON request body
	hashed := sha256.Sum256(body)
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
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
	pemBlock  *pem.Block
}

func (v *rsaVerifier) Verify(body []byte, signature string) error {
	// Decode signature from base64
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(body)
	err = rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hashed[:], decoded)
	if err != nil {
		return ErrInvalidRequestSignature(signature, err)
	}
	return nil
}

func (v *rsaVerifier) Key() []byte {
	buf := bytes.NewBuffer(nil)
	pem.Encode(buf, v.pemBlock)
	return buf.Bytes()
}

// NoopVerifier is useful to forgo
// all certificate verification.
type NoopVerifier struct{}

func (v NoopVerifier) Verify([]byte, string) error { return nil }

func (v NoopVerifier) Key() []byte { return []byte{} }

// NoopSigner is useful to forgo all
// signature generation.
type NoopSigner struct{}

func (s NoopSigner) Sign([]byte) (string, error) { return "", nil }
