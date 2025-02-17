package ntlm

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
)

// First message received by the server
func ServerNegociate(w http.ResponseWriter, r *http.Request) error {
	authorization := r.Header.Get("Authorization")
	authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
	if err != nil {
		errString := fmt.Errorf("ServerNegociate : error while decoding authorization header %s : %s\n", authorization, err)
		return errString
	}

	msg1 := NTLMSSP_NEGOTIATE{}
	msg1.Read(authorization_bytes)

	logger.Print("[+] NTLM NEGOCIATE\n")
	logger.Fprintf(logger.Writer(), "%s\n", msg1.ToString())

	return nil
}

// Second message received by the server
func ServerChallege(w http.ResponseWriter, r *http.Request, challenge string, domainName string) error {
	msg2 := NewNTLMSSP_CHALLENGEShort()
	// msg2 := NewNTLMSSP_CHALLENGE(challenge, domainName)
	msg2b64 := base64.RawStdEncoding.EncodeToString(msg2.ToBytes())

	header := fmt.Sprintf("NTLM %s", msg2b64)

	w.Header().Set("WWW-Authenticate", header)
	w.WriteHeader(401)

	return nil
}

// Third message received by the server
func ServerAuthenticate(w http.ResponseWriter, r *http.Request) (NTLMSSP_AUTH, NTLMv2Response, error) {
	authorization := r.Header.Get("Authorization")
	authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
	if err != nil {
		errString := fmt.Errorf("ServerAuthenticate : error while decoding authorization header %s : %s\n", authorization, err)
		return NTLMSSP_AUTH{}, NTLMv2Response{}, errString
	}

	msg3 := NTLMSSP_AUTH{}
	msg3.Read(authorization_bytes)

	logger.Printf("[+] NTLM AUTHENTICATE:\nTarget Name: %s\nUsername: %s\nWorkstation: %s\n", msg3.TargetName.RawData, msg3.Username.RawData, msg3.Workstation.RawData)
	fmt.Fprintf(logger.Writer(), "%s\n", msg3.ToString())

	// Prepare final response to the client
	ntlmv2Response := NTLMv2Response{}
	ntlmv2Response.Read(msg3.NTLMv2Response.RawData)

	fmt.Fprintf(logger.Writer(), "[+] NTLM AUTHENTICATE RESPONSE:\n%s\n", ntlmv2Response.ToString())

	return msg3, ntlmv2Response, nil
}

func NTLMPreliminaryChecks(w http.ResponseWriter, r *http.Request) error {
	authorization := r.Header.Get("Authorization")

	// Sometimes even if NTLM auth is required, the server is sending and other header
	if !strings.HasPrefix(authorization, "NTLM ") {
		err := fmt.Errorf("[NON NTLM HEADER CAPTURED] [%s]: %s\n", utils.GetSourceIP(r), authorization)
		return err
	}

	// Remove the "NTLM " string at the beginning
	authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
	if err != nil {
		err = fmt.Errorf("Decode error authorization header : %s : %s\n", authorization, err)
		return err
	}

	if len(authorization_bytes) < 40 {
		err := fmt.Errorf("Decoded authorization header is less than 40 bytes. Header was : %s\n", authorization)
		return err
	}

	if len(authorization_bytes) < 12 {
		err := fmt.Errorf("Decoded authorization header is less than 12 bytes. Header was : %s\n", authorization)
		return err
	}

	return nil
}

func NTLMDispatch(n NTLMAuth, w http.ResponseWriter, r *http.Request) (*NTLMSSP_AUTH, *NTLMv2Response, error) {
	authorization := r.Header.Get("Authorization")
	authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
	if err != nil {
		err := fmt.Errorf("Decode error authorization header : %s\n", authorization)
		return nil, nil, err
	}
	msgType := binary.LittleEndian.Uint32(authorization_bytes[8:12])

	// Received Negociate message. Handle it and answer with a Challenge message
	if msgType == uint32(1) {
		err := ServerNegociate(w, r)
		if err != nil {
			return nil, nil, err
		}

		err = ServerChallege(w, r, n.Challenge, n.DomainName)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, nil
	}

	// Retrieve information into the Authentication message
	if msgType == uint32(3) {
		msg3, ntlmv2Response, err := ServerAuthenticate(w, r)
		if err != nil {
			return nil, nil, err
		}
		return &msg3, &ntlmv2Response, err
	}

	return nil, nil, nil
}
