package gopserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	basicAuth "github.com/hophouse/gop/authentication/basic"
	ntlmAuth "github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

type Server struct {
	Host   string
	Port   string
	Scheme string
	Vhost  string
	Auth   string
	Realm  string
}

type ServerInterface interface {
	GetCertSubject() string
	CreateRouter() *mux.Router
	CreateMiddleware() *negroni.Negroni
	GetServer(*mux.Router, *negroni.Negroni) (*http.Server, error)
}

func (s Server) GetCertSubject() string {
	if s.Vhost != "" {
		return s.Vhost
	} else {
		return s.Host
	}
}

func (s Server) CreateMiddleware() *negroni.Negroni {
	n := negroni.New(negroni.NewRecovery())
	n.Use(&logMiddleware{})

	// Apply an auth system if requested
	switch strings.ToLower(s.Auth) {
	case "basic":
		logger.Printf("[+] Add HTTP Basic auth header\n")
		n.Use(&basicAuth.BasicAuthMiddleware{
			Realm: s.Realm,
		})
	case "ntlm":
		logger.Printf("[+] Add HTTP NTLM auth header\n")
		ntlmAuth.NtlmCapturedAuth = make(map[string]bool)
		n.Use(&ntlmAuth.NTLMAuthMiddleware{})
	}

	return n
}

func RunServerHTTPCmd(object ServerInterface) error {
	begin := time.Now()

	r := object.CreateRouter()
	n := object.CreateMiddleware()

	server, err := object.GetServer(r, n)
	if err != nil {
		return err
	}

	logger.Fatal(server.ListenAndServe())

	end := time.Now()
	logger.Printf("\n -  Execution time: %s\n", end.Sub(begin))

	return nil
}

func RunServerHTTPSCmd(object ServerInterface) error {
	begin := time.Now()

	r := object.CreateRouter()
	n := object.CreateMiddleware()

	server, err := object.GetServer(r, n)
	if err != nil {
		return err
	}

	caManager, err := gopproxy.InitCertManager("", "")
	if err != nil {
		logger.Fatal(err.Error())
	}

	cert, err := caManager.CreateCertificate(object.GetCertSubject())
	if err != nil {
		logger.Fatal(err.Error())
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

func RunRedirectServerHTTPCmd(host string, port string, vhost string, destination string, https bool) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.RedirectHandler(destination, http.StatusFound))

	n := negroni.New(negroni.NewRecovery())
	n.Use(&logMiddleware{})
	n.UseHandler(r)

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	if !https {
		logger.Printf("[+] Starting redirect server listening to : http://%s\n", addr)

		logger.Fatal(server.ListenAndServe())
	} else {
		caManager, err := gopproxy.InitCertManager("", "")
		if err != nil {
			logger.Fatal(err.Error())
		}

		certSubject := host
		if vhost != "" {
			certSubject = vhost
		}
		cert, err := caManager.CreateCertificate(certSubject)
		if err != nil {
			logger.Fatal(err.Error())
		}

		server.TLSConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}

		logger.Printf("[+] Starting redirect server listening to : https://%s\n", addr)
		logger.Fatal(server.ListenAndServeTLS("", ""))

	}

	return nil
}
