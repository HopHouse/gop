package ntlm

import (
	"fmt"
	"strings"
)

type Flag uint32

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
	0x02000000: "Unknown Flag",
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
