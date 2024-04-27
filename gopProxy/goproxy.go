package gopproxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
	"github.com/jroimartin/gocui"
)

func RunProxyCmd(options *Options) {
	// Init InterceptChan
	InterceptChan = make(chan bool, 1)

	// RunHTTPProxyCmd(options)
	RunNetProxyCmd(options)
}

func RunNetProxyCmd(options *Options) {
	addr := fmt.Sprintf("%s:%s", options.Host, options.Port)
	_, err := net.ResolveTCPAddr("tcp4", addr)
	utils.CheckErrorExit(err)

	certManager, err := InitCertManager(options.caFileOption, options.caPrivKeyFileOption)
	if err != nil {
		logger.Fatal(err)
	}

	l, err := net.Listen("tcp4", addr)
	utils.CheckErrorExit(err)
	defer l.Close()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				// handle error
				logger.Println(err)
				break
			}
			handleConnection(conn, &certManager)
		}
	}()

	err = RunGUI()
	if err != nil {
		logger.Printf("Error with RunGUI : %s\n", err)
	}
}

func handleConnection(conn net.Conn, certManager *CertManager) {
	defer conn.Close()
	// ClearAllGUIViews()

	// Copy request
	logger.Printf("Receive conenction from %s to %s\n", conn.RemoteAddr(), conn.LocalAddr())
	reader := bufio.NewReader(conn)

	req, err := http.ReadRequest(reader)
	if ok := utils.CheckError(err); ok {
		return
	}
	defer req.Body.Close()

	// Connect method
	if req.Method == "CONNECT" {

		// DNSLookup for IP
		_, err := net.ResolveTCPAddr("tcp4", req.URL.Host)
		if ok := utils.CheckError(err); ok {
			logger.Println(err)
			return
		}

		// Dial the client
		initConn, err := net.DialTimeout("tcp4", req.URL.Host, 2*time.Second)
		if ok := utils.CheckError(err); ok {
			logger.Println(err)
			return
		}
		defer initConn.Close()

		clientConn := tls.Client(initConn, &tls.Config{InsecureSkipVerify: true})
		err = clientConn.Handshake()
		if ok := utils.CheckError(err); ok {
			logger.Println(err)
			return
		}

		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\nProxy-agent: GoPentest/1.0\r\n\r\n"))
		if err != nil {
			logger.Println(err)
			return
		}

		cer, err := certManager.CreateCertificate(req.URL.Hostname())
		if err != nil {
			logger.Println(err)
			return
		}

		config := &tls.Config{
			Certificates:       []tls.Certificate{cer},
			InsecureSkipVerify: true,
		}

		proxyConn := tls.Server(conn, config)
		err = proxyConn.Handshake()
		if ok := utils.CheckError(err); ok {
			return
		}
		defer proxyConn.Close()

		proxyReader := bufio.NewReader(proxyConn)
		req, err := http.ReadRequest(proxyReader)
		if ok := utils.CheckError(err); ok {
			return
		}
		PrintGUIRequest(req)
		intercept()

		dumpedReq, _ := httputil.DumpRequest(req, true)
		_, err = clientConn.Write(dumpedReq)
		if err != nil {
			logger.Println(err)
			return
		}

		clientReader := bufio.NewReader(clientConn)
		res, err := http.ReadResponse(clientReader, req)
		if ok := utils.CheckError(err); ok {
			return
		}
		// defer res.Body.Close()

		PrintGUIResponse(*res)
		intercept()

		dumpedRes, _ := httputil.DumpResponse(res, true)
		_, err = proxyConn.Write(dumpedRes)
		if err != nil {
			logger.Println(err)
			return
		}

		proxyConn.Close()
		clientConn.Close()
		return
	}

	PrintGUIRequest(req)
	intercept()

	// Do request to target
	res := doHTTPRequest(req)
	if res == nil {
		return
	}
	// defer res.Body.Close()

	PrintGUIResponse(*res)
	intercept()

	sendNetResponse(conn, res)
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func sendNetResponse(conn net.Conn, resp *http.Response) {
	respBuffer, err := httputil.DumpResponse(resp, true)
	if ok := utils.CheckError(err); ok {
		return
	}

	_, err = conn.Write(respBuffer)
	if utils.CheckError(err) {
		return
	}
}

func copyHeader(newHeader http.Header, header http.Header) {
	for key, i := range header {
		for _, y := range i {
			newHeader.Add(key, y)
		}
	}
}

func doHTTPRequest(r *http.Request) *http.Response {
	newRequest, err := http.NewRequest(r.Method, r.URL.String(), nil)
	if ok := utils.CheckError(err); ok {
		return nil
	}

	copyHeader(newRequest.Header, r.Header)

	client := http.Client{}
	resp, err := client.Do(newRequest)
	if ok := utils.CheckError(err); ok {
		return nil
	}

	return resp
}

func PrintRequest(v *gocui.View, r *http.Request) {
	logger.Fprintf(v, "%s %s %s\n", r.Method, r.URL, r.Proto)
	logger.Fprintf(v, "Host: %s\n", r.Host)
	for headerName, headerValueSlice := range r.Header {
		for _, headerValue := range headerValueSlice {
			logger.Fprintf(v, "%s: %s\n", headerName, headerValue)
		}
	}
}

func PrintResponse(v *gocui.View, r http.Response) {
	logger.Fprintf(v, "%s\n", r.Status)
	for headerName, headerValueSlice := range r.Header {
		logger.Fprintf(v, "%s: %s\n", headerName, headerValueSlice[0])
	}
}

func PrintGUIRequest(r *http.Request) {
	if G == nil {
		return
	}

	G.Update(func(g *gocui.Gui) error {
		v := ClearGUIView(g, "host")
		logger.Fprintf(v, "%s", r.Host)

		v = ClearGUIView(g, "url")
		logger.Fprintf(v, "%s %s %s", r.URL.Scheme, r.URL.User, r.URL.Host)

		v = ClearGUIView(g, "request")
		PrintRequest(v, r)

		return nil
	})
}

func PrintGUIResponse(r http.Response) {
	if G == nil {
		return
	}

	// Read the content
	var bodyBytes []byte
	if r.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Request.Body)
	}
	// Restore the io.ReadCloser to its original state
	r.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Use the content

	body := string(bodyBytes)

	G.Update(func(g *gocui.Gui) error {
		v := ClearGUIView(g, "response-header")
		PrintResponse(v, r)

		v = ClearGUIView(g, "response-body")
		logger.Fprintf(v, "%s", body)
		return nil
	})
}

func intercept() {
	if InterceptMode {
		// Wait for data in channel and consimme it
		<-InterceptChan
	}
}
