package gopproxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/hophouse/gop/utils"
)

type Proxy struct {
	certManager CertManager
}

func RunHTTPProxyCmd(options *Options) {
	addr := fmt.Sprintf("%s:%s", options.Host, options.Port)
	_, err := net.ResolveTCPAddr("tcp4", addr)
	utils.CheckErrorExit(err)

	_, err = net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		utils.Log.Fatal(err)
		return
	}

	certManager, err := InitCertManager(options.caFileOption, options.caPrivKeyFileOption)
	if err != nil {
		utils.Log.Fatal(err)
		return
	}

	proxy := &Proxy{
		certManager: certManager,
	}

	server := &http.Server{Addr: addr, Handler: proxy}
	go server.ListenAndServe()
	//RunGUI(server)
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Connect method
	if req.Method == "CONNECT" {
		err := p.handleHTTPSMethod(w, req)
		if err != nil {
			utils.Log.Println(err)
			return
		}

		return
	}

	// Clean req URL

	PrintGUIRequest(req)
	intercept()

	// Do request to target
	req.Response = doHTTPRequest(req)
	var res *http.Response
	res = req.Response
	if res == nil {
		return
	}

	PrintGUIResponse(*res)
	intercept()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		utils.Log.Println(err)
		return
	}
	res.Body.Close()
	w.Write(body)
}

func (p Proxy) handleHTTPSMethod(w http.ResponseWriter, req *http.Request) error {
	// DNSLookup for IP
	_, err := net.ResolveTCPAddr("tcp4", req.URL.Host)
	if ok := utils.CheckError(err); ok {
		return err
	}

	// Dial the client
	clientConn, err := tls.Dial("tcp4", req.URL.Host, &tls.Config{InsecureSkipVerify: true})
	if ok := utils.CheckError(err); ok {
		return err
	}
	defer clientConn.Close()

	utils.Log.Printf("CONNECT from %s to %s \n", clientConn.LocalAddr(), clientConn.RemoteAddr())
	w.Write([]byte("HTTP/1.1 200 OK\r\nProxy-agent: GoPentest/1.0\r\n\r\n"))

	// Hijack the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return nil
	}
	conn, _, err := hj.Hijack()
	if ok := utils.CheckError(err); ok {
		return err
	}

	cer, err := p.certManager.CreateCertificate(req.URL.Hostname())
	if err != nil {
		return err
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cer},
		InsecureSkipVerify: true,
	}

	proxyConn := tls.Server(conn, config)
	err = proxyConn.Handshake()
	if ok := utils.CheckError(err); ok {
		return err
	}
	defer proxyConn.Close()

	proxyReader := bufio.NewReader(proxyConn)
	req, err = http.ReadRequest(proxyReader)
	if ok := utils.CheckError(err); ok {
		ClearAllGUIViews()
		return err
	}

	PrintGUIRequest(req)
	intercept()

	dumpedReq, _ := httputil.DumpRequest(req, true)
	clientConn.Write(dumpedReq)

	clientReader := bufio.NewReader(clientConn)
	res, err := http.ReadResponse(clientReader, req)
	if ok := utils.CheckError(err); ok {
		return err
	}
	if res == nil {
		return errors.New("Reponse is nill")
	}

	PrintGUIResponse(*res)
	intercept()

	// Clear all the views
	ClearAllGUIViews()

	dumpedRes, _ := httputil.DumpResponse(res, true)
	proxyConn.Write(dumpedRes)

	clientConn.Close()
	proxyConn.Close()

	return nil
}
