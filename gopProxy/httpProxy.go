package gopproxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
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
		logger.Fatal(err)
		return
	}

	certManager, err := InitCertManager(options.caFileOption, options.caPrivKeyFileOption)
	if err != nil {
		logger.Fatal(err)
		return
	}

	proxy := &Proxy{
		certManager: certManager,
	}

	server := &http.Server{Addr: addr, Handler: proxy}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalln("Error during server.ListenAndServe function")
		}
	}()
	// RunGUI(server)
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Connect method
	if req.Method == "CONNECT" {
		err := p.handleHTTPSMethod(w, req)
		if err != nil {
			logger.Println(err)
			return
		}

		return
	}

	// Clean req URL

	PrintGUIRequest(req)
	intercept()

	// Do request to target
	req.Response = doHTTPRequest(req)
	res := req.Response
	if res == nil {
		return
	}

	PrintGUIResponse(*res)
	intercept()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Println(err)
		return
	}

	res.Body.Close()
	_, err = w.Write(body)
	if err != nil {
		logger.Println(err)
		return
	}
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

	logger.Printf("CONNECT from %s to %s \n", clientConn.LocalAddr(), clientConn.RemoteAddr())
	_, err = w.Write([]byte("HTTP/1.1 200 OK\r\nProxy-agent: GoPentest/1.0\r\n\r\n"))
	if err != nil {
		return err
	}

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
	_, err = clientConn.Write(dumpedReq)
	if err != nil {
		return err
	}

	clientReader := bufio.NewReader(clientConn)
	res, err := http.ReadResponse(clientReader, req)
	if ok := utils.CheckError(err); ok {
		return err
	}
	if res == nil {
		return errors.New("reponse is nill")
	}

	PrintGUIResponse(*res)
	intercept()

	// Clear all the views
	ClearAllGUIViews()

	dumpedRes, _ := httputil.DumpResponse(res, true)
	_, err = proxyConn.Write(dumpedRes)
	if err != nil {
		return err
	}

	clientConn.Close()
	proxyConn.Close()

	return nil
}
