package gopserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	basicAuth "github.com/hophouse/gop/authentication/basic"
	ntlmAuth "github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

func genRedirectProxy(dstUrl string, auth string, realm string, redirectPrefix string) *negroni.Negroni {
	remote, err := url.Parse(dstUrl)
	if err != nil {
		panic(err)
	}

	// Router
	r := mux.NewRouter()

	proxy := &RedirectProxy{
		p:      httputil.NewSingleHostReverseProxy(remote),
		remote: remote,
	}

	if redirectPrefix != "" {
		route := fmt.Sprintf("/%s/", redirectPrefix)
		s := r.PathPrefix(route).Subrouter()
		s.HandleFunc("/{URL:.*}", proxy.HandleFunc)
	} else {
		r.HandleFunc("/{URL:.*}", proxy.HandleFunc)
	}

	n := negroni.New(negroni.NewRecovery())

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

	n.Use(XRealIPMiddleware{})
	n.UseHandler(r)

	return n
}

func RunRedirectProxyHTTPCmd(host string, port string, dstUrl string, auth string, realm string, redirectPrefix string) {
	begin := time.Now()

	addr := fmt.Sprintf("%s:%s", host, port)
	logger.Printf("[+] Starting reverse proxy listening to : http://%s and redirect to %s\n", addr, dstUrl)

	n := genRedirectProxy(dstUrl, auth, realm, redirectPrefix)

	logger.Fatal(http.ListenAndServe(addr, n))

	end := time.Now()
	logger.Printf("\n -  Execution time: %s\n", end.Sub(begin))
}

func RunRedirectProxyHTTPSCmd(host string, port string, dstUrl string, auth string, realm string, redirectPrefix string) {
	begin := time.Now()

	addr := fmt.Sprintf("%s:%s", host, port)

	n := genRedirectProxy(dstUrl, auth, realm, redirectPrefix)

	caManager, err := gopproxy.InitCertManager("", "")
	if err != nil {
		logger.Fatalf(err.Error())
	}

	cert, err := caManager.CreateCertificate(host)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	server := &http.Server{Addr: addr, Handler: n}
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	logger.Printf("[+] Starting reverse proxy listening to : https://%s and redirect to %s\n", addr, dstUrl)

	logger.Fatal(server.ListenAndServeTLS("", ""))

	end := time.Now()
	logger.Printf("\n -  Execution time: %s\n", end.Sub(begin))
}

type RedirectProxy struct {
	p      *httputil.ReverseProxy
	remote *url.URL
}

func (rp *RedirectProxy) HandleFunc(w http.ResponseWriter, r *http.Request) {
	// logger.Println("# %#v\n", r)

	// vars := mux.Vars(r)
	// value := vars["URL"]

	// logger.Println("# %#v\n", vars)

	// if value != "" {
	// 	redirectUrl, err := url.Parse(value)
	// 	if err != nil {
	// 		logger.Println("Could not parse URL to", vars["URL"])
	// 		return
	// 	}

	// 	r.URL = redirectUrl
	// } else {
	// 	r.URL, _ = url.Parse("/")
	// }

	r.Host = rp.remote.String()
	rp.p.ServeHTTP(w, r)
}
