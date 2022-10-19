package gopserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	basicAuth "github.com/hophouse/gop/authentication/basic"
	ntlmAuth "github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

func RunServerHTTPCmd(host string, port string, directory string, auth string, realm string) error {
	begin := time.Now()

	server, err := GetServerCmd(host, port, directory, auth, realm)
	if err != nil {
		return err
	}
	logger.Fatal(server.ListenAndServe())

	end := time.Now()
	logger.Printf("\n -  Execution time: %s\n", end.Sub(begin))

	return nil
}

func RunServerHTTPSCmd(host string, port string, directory string, auth string, realm string) error {
	begin := time.Now()

	server, err := GetServerCmd(host, port, directory, auth, realm)
	if err != nil {
		return nil
	}
	caManager, err := gopproxy.InitCertManager("", "")
	if err != nil {
		logger.Fatalf(err.Error())
	}

	cert, err := caManager.CreateCertificate(host)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	server.TLSConfig = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	logger.Fatal(server.ListenAndServeTLS("", ""))

	end := time.Now()
	logger.Printf("\n -  Execution time: %s\n", end.Sub(begin))

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
	logger.Printf("[+] Serve file to: http://%s for %s\n", addr, directory)

	// Router
	r := mux.NewRouter()

	fileServer := http.FileServer(http.Dir(directory))
	r.PathPrefix("/").Handler(fileServer)
	n := negroni.New(negroni.NewRecovery())
	n.Use(&logMiddleware{})

	// Apply an auth system if requested
	switch strings.ToLower(auth) {
	case "basic":
		logger.Printf("[+] Add HTTP Basic auth header\n")
		n.Use(&basicAuth.BasicAuthMiddleware{
			Realm: realm,
		})
	case "ntlm":
		logger.Printf("[+] Add HTTP NTLM auth header\n")
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
