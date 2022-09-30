package gopRelay

import (
	"net/http"
	"net/http/httputil"

	"github.com/hophouse/gop/authentication/ntlm"
	"github.com/hophouse/gop/utils/logger"
)

var client map[string]bool

func Run() {
	addr := "127.0.0.1:8080"

	logger.Printf("[+] Run server on : %s\n", addr)

	// Create a server to listen for requests
	err := http.ListenAndServe(addr, &ntlm.NTLMAuthMiddleware{
		Challenge:     "00000000",
		DomainName:    "smbdomain",
		ServerName:    "DC",
		DnsDomainName: "smbdomain.local",
		DnsServerName: "dc.smbdomain.local",
	})
	if err != nil {
		logger.Fatal(err)
	}
}

func HandlerFunc(w http.ResponseWriter, r *http.Request) {

	// Once a request is received use it to relay it to a website
	logger.Printf("[+] Received a new connexion from : %s\n", r.RemoteAddr)
	_, exist := client[r.RemoteAddr]
	if !exist {
		logger.Printf("[+] Received a new connexion from : %s\n", r.RemoteAddr)
	}

	initialRequestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.Println(err)
		return
	}
	logger.Println(string(initialRequestDump))

	err = ntlm.ServerNegociate(w, r)
	if err != nil {
		logger.Println(err)
		return
	}
}
