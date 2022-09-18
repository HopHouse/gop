package gopserver

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	basicAuth "github.com/hophouse/gop/auth/basic"
	ntlmAuth "github.com/hophouse/gop/auth/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils"
	"github.com/urfave/negroni"
)

func RunServerHTTPCmd(host string, port string, directory string, auth string, realm string) error {
	begin := time.Now()

	server, err := GetServerCmd(host, port, directory, auth, realm)
	if err != nil {
		return err
	}
	utils.Log.Fatal(server.ListenAndServe())

	end := time.Now()
	fmt.Printf("\n -  Execution time: %s\n", end.Sub(begin))

	return nil
}

func RunServerHTTPSCmd(host string, port string, directory string, auth string, realm string) error {
	begin := time.Now()

	server, err := GetServerCmd(host, port, directory, auth, realm)
	if err != nil {
		return nil
	}
	serverCert, serverKey := gopproxy.GenerateCA()

	caBytes, err := x509.CreateCertificate(rand.Reader, serverCert, serverCert, serverKey.Public(), serverKey)
	if err != nil {
		fmt.Println(err)
		utils.Log.Fatal(err)
	}

	serverCertPEM := new(bytes.Buffer)
	pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	serverPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})

	cer, err := tls.X509KeyPair(serverCertPEM.Bytes(), serverPrivKeyPEM.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	server.TLSConfig = &tls.Config{
		Certificates:       []tls.Certificate{cer},
		InsecureSkipVerify: true,
	}
	utils.Log.Fatal(server.ListenAndServeTLS("", ""))

	end := time.Now()
	fmt.Printf("\n -  Execution time: %s\n", end.Sub(begin))

	return nil
}

func GetServerCmd(host string, port string, directory string, auth string, realm string) (http.Server, error) {
	path, err := os.Getwd()
	if err != nil {
		return http.Server{}, err
	}

	if !strings.HasPrefix(directory, "/") && !strings.HasPrefix(directory, "C:\\") {
		directory = filepath.Join(path, directory)
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("[+] Serve file to: http://%s for %s\n", addr, directory)

	// Router
	r := mux.NewRouter()

	fileServer := http.FileServer(http.Dir(directory))
	r.PathPrefix("/").Handler(fileServer)
	n := negroni.New(negroni.NewRecovery())
	n.Use(&logMiddleware{})

	// Apply an auth system if requested
	switch strings.ToLower(auth) {
	case "basic":
		fmt.Printf("[+] Add HTTP Basic auth header\n")
		n.Use(&basicAuth.BasicAuthMiddleware{
			Realm: realm,
		})
	case "ntlm":
		fmt.Printf("[+] Add HTTP NTLM auth header\n")
		ntlmAuth.NtlmCapturedAuth = make(map[string]bool)
		n.Use(&ntlmAuth.NTLMAuthMiddleware{})
	}

	n.UseHandler(r)

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	return server, nil
}
