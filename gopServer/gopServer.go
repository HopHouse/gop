package gopserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	basicAuth "github.com/hophouse/gop/auth/basic"
	ntlmAuth "github.com/hophouse/gop/auth/ntlm"
	"github.com/hophouse/gop/utils"
	"github.com/urfave/negroni"
)

func RunServerHTTPCmd(host string, port string, directory string, auth string, realm string) {
	begin := time.Now()
	path, err := os.Getwd()
	if err != nil {
		utils.Log.Println(err)
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
	utils.Log.Fatal(http.ListenAndServe(addr, n))

	end := time.Now()
	fmt.Printf("\n -  Execution time: %s\n", end.Sub(begin))
}
