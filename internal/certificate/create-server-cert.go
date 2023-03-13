package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
)

func CreateServerCert() tls.Certificate {
	caCertPem, caPrivKeyPem, _, _ := CreateCA()
	certCertPem, certPrivKeyPem, certPublicKeyPem, _ := CreateCert()

	caCertBlob, _ := pem.Decode(caCertPem)
	caCert, _ := x509.ParseCertificate(caCertBlob.Bytes)
	caPrivKeyBlob, _ := pem.Decode(caPrivKeyPem)
	caPrivKey, _ := x509.ParsePKCS1PrivateKey(caPrivKeyBlob.Bytes)
	certCertBlob, _ := pem.Decode(certCertPem)
	certCert, _ := x509.ParseCertificate(certCertBlob.Bytes)
	certPublicKeyBlob, _ := pem.Decode(certPublicKeyPem)
	certPublicKey, _ := x509.ParsePKCS1PublicKey(certPublicKeyBlob.Bytes)

	signedCert, err := x509.CreateCertificate(rand.Reader, certCert, caCert, certPublicKey, caPrivKey)
	if err != nil {
		log.Fatalf("[FATAL] CreateCertificate: %v", err)
	}

	serverCertPem := new(bytes.Buffer)
	pem.Encode(serverCertPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: signedCert,
	})

	serverCert, err := tls.X509KeyPair(append(serverCertPem.Bytes(), caCertPem...), certPrivKeyPem)
	if err != nil {
		log.Fatalf("[FATAL] x509KeyPair: %v", err)
	}

	return serverCert
}
