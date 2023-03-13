package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strconv"
	"time"
)

func CreateCert() ([]byte, []byte, []byte, error) {
	serial, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(serial + 1),
		Subject: pkix.Name{
			Organization: []string{"Kubernetes Mutating Webserver"},
			Country:      []string{"K8S"},
			Province:     []string{"Cluster Service"},
			Locality:     []string{"Cluster Local"},
			//StreetAddress: []string{""},
			//PostalCode:    []string{""},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 6, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames: []string{
			"svc.cluster.local",
			"*.svc.cluster.local",
		},
		SubjectKeyId:       []byte{1, 2, 3, 4, 6},
		KeyUsage:           x509.KeyUsageDigitalSignature,
		SignatureAlgorithm: x509.SHA384WithRSA,
	}

	keyPair, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return []byte(""), []byte(""), []byte(""), err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &keyPair.PublicKey, keyPair)
	if err != nil {
		return []byte(""), []byte(""), []byte(""), err
	}

	c := new(bytes.Buffer)
	pem.Encode(c, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	k := new(bytes.Buffer)
	pem.Encode(k, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
	})

	p := new(bytes.Buffer)
	pem.Encode(p, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&keyPair.PublicKey),
	})

	return c.Bytes(), k.Bytes(), p.Bytes(), nil
}
