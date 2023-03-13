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

func CreateCA() ([]byte, []byte, []byte, error) {
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
