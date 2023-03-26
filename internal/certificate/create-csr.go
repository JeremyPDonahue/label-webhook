package certificate

import (
	"bytes"
	"fmt"
	"log"

	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
)

func CreateCSR(privateKey string, dnsNames []string) (string, error) {
	dnsNames = append(dnsNames, "*.svc.cluster.local")

	csr := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"Kubernetes Mutating Webserver"},
			Country:      []string{"K8S"},
			Province:     []string{"Cluster Service"},
			Locality:     []string{"Cluster Local"},
			//StreetAddress: []string{""},
			//PostalCode:    []string{""},
		},
		DNSNames: dnsNames,
		SignatureAlgorithm: x509.SHA384WithRSA,
	}

	pemKey, _ := pem.Decode([]byte(privateKey))
	if pemKey == nil || pemKey.Type != "RSA PRIVATE KEY" {
		return "", fmt.Errorf("failed to decode PEM block containing private key")
	}
	keyPair, err := x509.ParsePKCS1PrivateKey(pemKey.Bytes)
	if err != nil {
		return "", err
	}

	csrData, err := x509.CreateCertificateRequest(rand.Reader, &csr, keyPair)
	if err != nil {
		return "", err
	}

	c := new(bytes.Buffer)
	pem.Encode(c, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrData,
	})

	log.Printf("[TRACE] Generated Host CSR:\n%s", c.String())
	return c.String(), nil
}
