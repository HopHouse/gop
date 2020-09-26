package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type SecurityBuffer struct {
	BufferLength          uint16
	BufferAllocatedLength uint16
	StartOffset           uint32
}

func NewSecurityBuffer(data []byte) SecurityBuffer {
	buffer := SecurityBuffer{
		BufferLength:          binary.LittleEndian.Uint16(data[0:2]),
		BufferAllocatedLength: binary.LittleEndian.Uint16(data[3:5]),
		StartOffset:           binary.LittleEndian.Uint32(data[5:9])}
	return buffer
}

func (buf SecurityBuffer) PrintSecurityBuffer() {
	fmt.Printf("\tBuffer length                : %x\n", buf.BufferLength)
	fmt.Printf("\tBuffer Allocated length      : %x\n", buf.BufferAllocatedLength)
	fmt.Printf("\tOffset                       : %x\n", buf.BufferLength)
}

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
	0x00004000: "Unknown Local Call",
	0x00008000: "Negotiate Always Sign",
	0x00010000: "Target Type Domain",
	0x00020000: "Target Type Server",
	0x00040000: "Targeg Type Share",
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

func (flag Flag) PrintFlags() {
	for key, value := range flags {
		if key&flag != 0 {
			fmt.Printf("\t%s\n", value)
		}
	}
}

func (flag Flag) SetFlag(setFlag string) {
	for key, value := range flags {
		if strings.ToLower(value) == strings.ToLower(setFlag) {
			flag = flag & key
		}
	}
}

type NTLMType1 struct {
	SSPSignature        []byte
	MessageType         uint32
	Flags               Flag
	SuppliedDomain      []byte
	SuppliedWorkstation []byte
	OSVersionStructure  []byte
	OtherData           []byte
}

func NewNTLMType1(data []byte) *NTLMType1 {
	suppliedDomainBuffer := NewSecurityBuffer(data[16:24])
	var suppliedDomainBytes []byte
	if suppliedDomainBuffer.BufferAllocatedLength > 0 {
		suppliedDomainBytes = data[suppliedDomainBuffer.StartOffset : suppliedDomainBuffer.BufferAllocatedLength+1]
	}
	suppliedWorkstationBuffer := NewSecurityBuffer(data[24:33])
	var suppliedWorkstationBytes []byte
	if suppliedWorkstationBuffer.BufferAllocatedLength > 0 {
		suppliedWorkstationBytes = data[suppliedWorkstationBuffer.StartOffset : suppliedWorkstationBuffer.BufferAllocatedLength+1]
	}

	msg := NTLMType1{
		SSPSignature:        data[0:8],
		MessageType:         binary.LittleEndian.Uint32(data[8:12]),
		Flags:               (Flag)(binary.BigEndian.Uint32(data[12:16])),
		SuppliedDomain:      suppliedDomainBytes,
		SuppliedWorkstation: suppliedWorkstationBytes,
		OSVersionStructure:  data[32:41],
		OtherData:           data[40:],
	}
	return &msg
}

func (msg *NTLMType1) PrintNTLNType1() {
	fmt.Printf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature))
	fmt.Printf("NTLM Message Type      : %v\n", msg.MessageType)
	fmt.Printf("Flags : \n")
	msg.Flags.PrintFlags()
	fmt.Printf("Supplied Domain : %x\n", msg.SuppliedDomain)
	fmt.Printf("Supplied Workstation : %x\n", msg.SuppliedWorkstation)
	fmt.Printf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4]))
	fmt.Printf("\n")
}

type NTLMType2 struct {
}
