package certificate

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
)

func SignCert(caCertPem, caPrivKeyPem, csrPem string) (string, error) {
	caCertData, _ := pem.Decode([]byte(caCertPem))
	caCert, err := x509.ParseCertificate(caCertData.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse cert %v", err)
	}

	pemKey, _ := pem.Decode([]byte(caPrivKeyPem))
	if pemKey == nil || pemKey.Type != "RSA PRIVATE KEY" {
		return "", fmt.Errorf("failed to decode PEM block containing private key")
	}
	keyPair, err := x509.ParsePKCS1PrivateKey(pemKey.Bytes)
	if err != nil {
		return "", fmt.Errorf("private key %v", err)
	}

	csrData, _ := pem.Decode([]byte(csrPem))
	csr, err := x509.ParseCertificateRequest(csrData.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse csr %v", err)
	}

	serial, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(serial + 1),
		Issuer:       caCert.Issuer,
		Subject: pkix.Name{
			Organization:  []string{"Kubernetes Mutating Webserver"},
			Country:       csr.Subject.Country,
			Province:      []string{"Cluster Service"},
			Locality:      []string{"Cluster Local"},
			StreetAddress: csr.Subject.StreetAddress,
			PostalCode:    csr.Subject.PostalCode,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 6, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames:           csr.DNSNames,
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,
		PublicKey:          csr.PublicKey,
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, csr.PublicKey, keyPair)
	if err != nil {
		return "", fmt.Errorf("sign %v", err)
	}
	c := new(bytes.Buffer)
	pem.Encode(c, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	return c.String(), nil
}
