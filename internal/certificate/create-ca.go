package certificate

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
)

func CreateCA(privateKey string) (string, error) {
	serial, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			Organization: []string{"Kubernetes Mutating Webserver CA"},
			Country:      []string{"K8S"},
			Province:     []string{"Cluster Service"},
			Locality:     []string{"Cluster Local"},
			//StreetAddress: []string{""},
			//PostalCode:    []string{""},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
		IsCA:      true,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA384WithRSA,
	}

	pemKey, _ := pem.Decode([]byte(privateKey))
	if pemKey == nil || pemKey.Type != "RSA PRIVATE KEY" {
		return "", fmt.Errorf("failed to decode PEM block containing private key")
	}
	keyPair, err := x509.ParsePKCS1PrivateKey(pemKey.Bytes)
	if err != nil {
		return "", err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &keyPair.PublicKey, keyPair)
	if err != nil {
		return "", err
	}

	c := new(bytes.Buffer)
	pem.Encode(c, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	log.Printf("[DEBUG] Generated Certificate Authority Certificate:\n%s", c)
	return c.String(), nil
}
