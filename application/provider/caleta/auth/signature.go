package auth

// Verifier verifies a signature with a payload
type Verifier interface {
	// Verifies the payload and returns nil if correct and error if not
	Verify(signature string, payload []byte) error
}

// Signer signs a payload
type Signer interface {
	// Sign signs the payload and returns a signature or error if it fails
	Sign(payload []byte) ([]byte, error)
}
