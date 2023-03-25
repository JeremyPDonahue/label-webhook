package certificate

import (
	"crypto/rand"
	"crypto/rsa"
)

func CreateRSAKeyPair(bytes int) (*rsa.PrivateKey, error) {
	keyPair, err := rsa.GenerateKey(rand.Reader, bytes)
	if err != nil {
		return &rsa.PrivateKey{}, err
	}

	return keyPair, nil
}
