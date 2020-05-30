package gopproxy

import (
	"time"
	"fmt"
	"io/ioutil"
	"crypto/tls"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"bytes"
	"net"

	"github.com/hophouse/gop/utils"
)

type CertManager struct {
	caCRT *x509.Certificate
	caPrivateKey *rsa.PrivateKey
	certPrivKey *rsa.PrivateKey
	certPrivKeyPEM *bytes.Buffer
	certStore map[string]tls.Certificate
}

func (certManager CertManager) CreateCertificate(host string) (tls.Certificate) {
	// Check if we already have this certificate
	if certificat, ok := certManager.certStore[host]; ok {
		return certificat
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		utils.Log.Fatalf("Failed to generate serial number: %v", err)
	}

    // certparameters
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:    host,
			Organization:  []string{"Company, INC."},
			Country:       []string{"FR"},
			Province:      []string{""},
			Locality:      []string{"Paris"},
			PostalCode:    []string{"75000"},
		},
		//IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	if ip := net.ParseIP(host); ip != nil {
		cert.IPAddresses = append(cert.IPAddresses, ip)
	} else {
		cert.DNSNames = append(cert.DNSNames, host)
	}

	// generate cert
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, certManager.caCRT, &certManager.certPrivKey.PublicKey, certManager.caPrivateKey)
	if err != nil {
		fmt.Println(err)
		utils.Log.Fatal(err)
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	//Save it
	//ioutil.WriteFile("server.crt", certPEM.Bytes(), 0644)

	cer, err := tls.X509KeyPair(certPEM.Bytes(), certManager.certPrivKeyPEM.Bytes())
	if err != nil {
		fmt.Println(err)
		return tls.Certificate{}
	}

	// Add the certificate to the store
	certManager.certStore[host] = cer

	return cer
}

func InitCertManager() CertManager {
	certManager := CertManager{}

	// Init store
	certManager.certStore = make(map[string]tls.Certificate)

	// load CA public key/certificate
	caPublicKeyFile, err := ioutil.ReadFile("../ca.crt")
	if err != nil {
		fmt.Println("Erreur reading CA")
		utils.Log.Fatal(err)
	}
	pemBlock, _ := pem.Decode(caPublicKeyFile)
	if pemBlock == nil {
		fmt.Println("Erreur decode CA")
		utils.Log.Fatal(err)
	}
	certManager.caCRT, err = x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		fmt.Println("Erreur parse CA")
		utils.Log.Fatal(err)
	}

	// Load CA Private key
	caPrivateKeyFile, err := ioutil.ReadFile("../ca.key")
	if err != nil {
		fmt.Println("Erreur reading CA key")
		utils.Log.Fatal(err)
	}
	pemBlock, _ = pem.Decode(caPrivateKeyFile)
	if pemBlock == nil {
		fmt.Println("Erreur decode CA key")
		utils.Log.Fatal(err)
	}
	certManager.caPrivateKey, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		fmt.Println(err)
		utils.Log.Fatal(err)
	}

	// Generate certPrivkey
	certManager.certPrivKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
		utils.Log.Fatal(err)
	}

	// Cert private key
	certManager.certPrivKeyPEM = new(bytes.Buffer)
	pem.Encode(certManager.certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certManager.certPrivKey),
	})
	//ioutil.WriteFile("server.key", certPrivKeyPEM.Bytes(), 0644)

	return certManager
}