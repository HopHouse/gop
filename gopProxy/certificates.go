package gopproxy

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"time"

	"github.com/hophouse/gop/utils"
)

type CertManager struct {
	CaCRT     *x509.Certificate
	CaCertPEM *bytes.Buffer

	CaPrivKey    *rsa.PrivateKey
	CaPrivKeyPem *bytes.Buffer

	CertPrivKey    *rsa.PrivateKey
	CertPrivKeyPEM *bytes.Buffer

	CertStore map[string]tls.Certificate
}

func (certManager CertManager) SaveKeysToDisk() error {
	var err error

	// CA Certificate
	err = ioutil.WriteFile("ca.crt", certManager.CaCertPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	// CA Private Key
	err = ioutil.WriteFile("ca-privkey.key", certManager.CaPrivKeyPem.Bytes(), 0644)
	if err != nil {
		return err
	}

	// Certificate Private Key
	err = ioutil.WriteFile("cert-privkey.key", certManager.CertPrivKeyPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	for name, cert := range certManager.CertStore {
		fileName := fmt.Sprintf("%s.crt", name)
		if cert.Certificate[0][:] != nil {
			// Certificate
			certPEM := new(bytes.Buffer)
			pem.Encode(certPEM, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Certificate[0][:],
			})
			err := ioutil.WriteFile(fileName, certPEM.Bytes(), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (certManager CertManager) CreateCertificate(host string) (tls.Certificate, error) {
	// Check if we already have this certificate
	if certificat, ok := certManager.CertStore[host]; ok {
		return certificat, nil
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		utils.Log.Fatalf("Failed to generate serial number: %v", err)
		return tls.Certificate{}, err
	}

	// certparameters
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"Company, INC."},
			Country:      []string{"FR"},
			Province:     []string{""},
			Locality:     []string{"Paris"},
			PostalCode:   []string{"75000"},
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
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, certManager.CaCRT, &certManager.CertPrivKey.PublicKey, certManager.CaPrivKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	//Save it
	//ioutil.WriteFile("server.crt", certPEM.Bytes(), 0644)

	cer, err := tls.X509KeyPair(certPEM.Bytes(), certManager.CertPrivKeyPEM.Bytes())
	if err != nil {
		return tls.Certificate{}, err
	}

	// Add the certificate to the store
	certManager.CertStore[host] = cer

	return cer, err
}

func InitCertManager(caFile string, caPrivKeyFile string) (CertManager, error) {
	var err error

	certManager := CertManager{}

	// Init store
	certManager.CertStore = make(map[string]tls.Certificate)

	// load CA public key/certificate
	if caFile != "" && caPrivKeyFile != "" {
		caPublicKeyFile, err := ioutil.ReadFile(caFile)
		if err != nil {
			err := fmt.Errorf("Error reading CA - %s", err)
			return CertManager{}, err
		}
		pemBlock, _ := pem.Decode(caPublicKeyFile)
		if pemBlock == nil {
			err := fmt.Errorf("Error decoding CA - %s", err)
			return CertManager{}, err
		}
		certManager.CaCRT, err = x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			err := fmt.Errorf("Error parsing CA - %s", err)
			return CertManager{}, err
		}

		// Load CA Private key
		caPrivateKeyFile, err := ioutil.ReadFile(caPrivKeyFile)
		if err != nil {
			err := fmt.Errorf("Error reading CA private key - %s", err)
			return CertManager{}, err
		}
		pemBlock, _ = pem.Decode(caPrivateKeyFile)
		if pemBlock == nil {
			err := fmt.Errorf("Error decoding CA private key - %s", err)
			return CertManager{}, err
		}

		certManager.CaPrivKey, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return CertManager{}, err
		}
	} else {
		certManager.CaCRT, certManager.CaPrivKey = GenerateCA()

		caBytes, err := x509.CreateCertificate(rand.Reader, certManager.CaCRT, certManager.CaCRT, certManager.CaPrivKey.Public(), certManager.CaPrivKey)
		if err != nil {
			fmt.Println(err)
			utils.Log.Fatal(err)
		}

		certManager.CaCertPEM = new(bytes.Buffer)
		pem.Encode(certManager.CaCertPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		})

		certManager.CaPrivKeyPem = new(bytes.Buffer)
		pem.Encode(certManager.CaPrivKeyPem, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(certManager.CaPrivKey),
		})
	}

	// Generate certPrivkey
	certManager.CertPrivKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return CertManager{}, err
	}

	// Cert private key
	certManager.CertPrivKeyPEM = new(bytes.Buffer)
	pem.Encode(certManager.CertPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certManager.CertPrivKey),
	})
	//ioutil.WriteFile("server.key", certPrivKeyPEM.Bytes(), 0644)

	return certManager, nil
}

func GenerateCA() (*x509.Certificate, *rsa.PrivateKey) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName:   "GOP",
			Organization: []string{"Company, INC."},
			Country:      []string{"FR"},
			Province:     []string{""},
			Locality:     []string{"Paris"},
			PostalCode:   []string{"75000"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, _ := rsa.GenerateKey(rand.Reader, 4096)

	return ca, caPrivKey
}
