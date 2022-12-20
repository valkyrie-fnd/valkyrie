package caleta

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
)

type signer struct {
	privateKey *rsa.PrivateKey
}

type verifier struct {
	publicKey rsa.PublicKey
}

// NewSigner accepts a PEM private rsa key and creates a new signer
func NewSigner(privateKey []byte) (auth.Signer, error) {
	block, _ := pem.Decode(privateKey)

	pKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &signer{pKey}, nil
}

// Sign creates a signature for the payload and encodes it using base64 encoding
func (s *signer) Sign(payload []byte) ([]byte, error) {
	hashed := sha256.Sum256(payload)
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}
	encodedSig := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(encodedSig, signature)
	return encodedSig, nil
}

// NewVerifier accepts a PEM public key and creates a new verifier
func NewVerifier(publicKey []byte) (auth.Verifier, error) {
	block, _ := pem.Decode(publicKey)
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		return &verifier{*key}, nil
	default:
		return nil, fmt.Errorf("unknown public key")
	}
}

// Verify decodes signature using base64 and verifies the payload. Returns nil if successful, otherwise error
func (v *verifier) Verify(signature string, payload []byte) error {
	hashed := sha256.Sum256(payload)
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(&v.publicKey, crypto.SHA256, hashed[:], decodedSignature)
}
