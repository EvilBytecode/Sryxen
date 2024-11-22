package Crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
	"net"
)

const (
	PrivateKeyPath = "Certs/keyfile.pem"
	PublicKeyPath  = "Certs/certfile.pem"
)

func GenerateRSAKeys() error {
	if err := os.MkdirAll("Certs", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create Certs directory: %v", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate RSA keys: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       []string{"Sryxen"},
			OrganizationalUnit: []string{"Sryxen-Server"},
			Locality:           []string{"Chicago"},
			Province:           []string{"Illinois"},
			Country:            []string{"US"},
			CommonName:         "sryxen.gg",
		},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<63-1))
	if err != nil {
		return fmt.Errorf("failed to generate random serial number: %v", err)
	}
	template.SerialNumber = serialNumber

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err := os.WriteFile(PrivateKeyPath, privPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key to file: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(PublicKeyPath, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to write certificate to file: %v", err)
	}

	fmt.Println("RSA keys and self-signed certificate generated and saved to Certs directory.")
	return nil
}
