package ntlm

import (
	"fmt"
	"strings"
)

type Flag uint32

const NTLMSSP_NEGOTIATE_UNICODE Flag = 0x00000001
const NTLMSSP_NEGOTIATE_OEM Flag = 0x00000002
const NTLMSSP_REQUEST_TARGET Flag = 0x00000004
const NTLMSSP_UNKNOWN00000008_Flag Flag = 0x00000008
const NTLMSSP_NEGOTIATE_SIGN Flag = 0x00000010
const NTLMSSP_NEGOTIATE_SEAL Flag = 0x00000020
const NTLMSSP_NEGOTIATE_DATAGRAM_Style Flag = 0x00000040
const NTLMSSP_NEGOTIATE_LAN_Manager_KEY Flag = 0x00000080
const NTLMSSP_NEGOTIATE_NETWARE Flag = 0x00000100
const NTLMSSP_NEGOTIATE_NTLM Flag = 0x00000200
const NTLMSSP_UNKNOWN00000400_FLAG Flag = 0x00000400
const NTLMSSP_NEGOTIATE_ANONYMOUS Flag = 0x00000800
const NTLMSSP_NEGOTIATE_DOMAIN_SUPPLIED Flag = 0x00001000
const NTLMSSP_NEGOTIATE_WORKSTATION_SUPPLIED Flag = 0x00002000
const NTLMSSP_NEGOTIATE_LOCAL_CALL Flag = 0x00004000
const NTLMSSP_NEGOTIATE_ALWAYS_SIGN Flag = 0x00008000
const NTLMSSP_TARGET_TYPE_DOMAIN Flag = 0x00010000
const NTLMSSP_TARGET_TYPE_SERVER Flag = 0x00020000
const NTLMSSP_TARGET_TYPE_SHARE Flag = 0x00040000
const NTLMSSP_NEGOTIATE_NTLMv2_KEY Flag = 0x00080000
const NTLMSSP_REQUEST_INIT_RESPONSE Flag = 0x00100000
const NTLMSSP_REQUEST_ACCEPT_RESPONSE Flag = 0x00200000
const NTLMSSP_REQUEST_NONNT_SESSION_KEY Flag = 0x00400000
const NTLMSSP_NEGOTIATE_TARGET_Info Flag = 0x00800000
const NTLMSSP_UNKNOWN01000000_Flag Flag = 0x01000000
const NTLMSSP_NEGOTIATE_VERSION Flag = 0x02000000
const NTLMSSP_UNKNOWN04000000_Flag Flag = 0x04000000
const NTLMSSP_UNKNOWN08000000_Flag Flag = 0x08000000
const NTLMSSP_UNKNOWN10000000_Flag Flag = 0x10000000
const NTLMSSP_NEGOTIATE_128 Flag = 0x20000000
const NTLMSSP_NEGOTIATE_KEY_EXCHANGE Flag = 0x40000000
const NTLMSSP_NEGOTIATE_56 Flag = 0x80000000

var flags = map[Flag]string{
	0x00000001: "Negotiate Unicode",
	0x00000002: "Negotiate OEM",
	0x00000004: "Request Target",
	0x00000008: "Unknown Flag",
	0x00000010: "Negotiate Sign",
	0x00000020: "Negotiate Seal",
	0x00000040: "Negotiate Datagram Style",
	0x00000080: "Negotiate Lan Manager Key",
	0x00000100: "Negotiate Netware",
	0x00000200: "Negotiate NTLM",
	0x00000400: "Unknown Flag",
	0x00000800: "Negotiate Anonymous",
	0x00001000: "Negotiate Domain Supplied",
	0x00002000: "Negotiate Workstation Supplied",
	0x00004000: "Negotiate Local Call",
	0x00008000: "Negotiate Always Sign",
	0x00010000: "Target Type Domain",
	0x00020000: "Target Type Server",
	0x00040000: "Target Type Share",
	0x00080000: "Negotiate NTLMv2 Key",
	0x00100000: "Request Init Response",
	0x00200000: "Request Accept Response",
	0x00400000: "Request Non-NT Session Key",
	0x00800000: "Negotiate Target Info",
	0x01000000: "Unknown Flag",
	0x02000000: "Negotiate Version",
	0x04000000: "Unknown Flag",
	0x08000000: "Unknown Flag",
	0x10000000: "Unknown Flag",
	0x20000000: "Negotiate 128",
	0x40000000: "Negotiate Key Exchange",
	0x80000000: "Negotiate 56",
}

func (flag *Flag) ToString() string {
	var str strings.Builder

	for key, value := range flags {
		if key&*flag != 0 {
			str.WriteString(fmt.Sprintf("\t%s\n", value))
		}
	}

	return str.String()
}

func (flag *Flag) SetFlag(setFlags ...string) {
	for _, setFlag := range setFlags {
		for key, value := range flags {
			if strings.EqualFold(value, setFlag) {
				*flag = *flag | key
				break
			}
		}
	}
}
