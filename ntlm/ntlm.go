package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const challenge string = "HopHouse"
const domainName string = "smbdomain"
const serverName string = "DC"
const dnsDomainName string = "smbdomain.local"
const dnsServerName string = "dc.smbdomain.local"

type SecurityBuffer struct {
	BufferLength          uint16
	BufferAllocatedLength uint16
	StartOffset           uint32
	RawData               []byte
}

func NewSecurityBuffer() SecurityBuffer {
	return SecurityBuffer{
		BufferLength:          uint16(0),
		BufferAllocatedLength: uint16(0),
		StartOffset:           uint32(0),
		RawData:               []byte{},
	}
}

func ReadSecurityBuffer(data []byte) SecurityBuffer {
	buffer := SecurityBuffer{
		BufferLength:          binary.LittleEndian.Uint16(data[0:2]),
		BufferAllocatedLength: binary.LittleEndian.Uint16(data[3:5]),
		StartOffset:           binary.LittleEndian.Uint32(data[5:9]),
		RawData:               []byte{},
	}

	if buffer.BufferAllocatedLength > 0 {
		buffer.RawData = data[int(buffer.StartOffset) : int(buffer.BufferAllocatedLength)+1]
	}

	return buffer
}

func (sbuf SecurityBuffer) ToBytes() []byte {
	//buffer := make([]byte, 0, 8)
	buffer := make([]byte, 0, 0)

	bufferLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bufferLengthBytes, sbuf.BufferLength)
	buffer = append(buffer, bufferLengthBytes...)

	bufferAllocatedLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bufferAllocatedLengthBytes, sbuf.BufferAllocatedLength)
	buffer = append(buffer, bufferAllocatedLengthBytes...)

	bufferStartOffsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bufferStartOffsetBytes, sbuf.StartOffset)
	buffer = append(buffer, bufferStartOffsetBytes...)

	return buffer
}

func (sbuf *SecurityBuffer) PrintSecurityBuffer() {
	fmt.Printf("\tBuffer length                : %x\n", sbuf.BufferLength)
	fmt.Printf("\tBuffer Allocated length      : %x\n", sbuf.BufferAllocatedLength)
	fmt.Printf("\tOffset                       : %x\n", sbuf.StartOffset)
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

func (flag *Flag) PrintFlags() {
	for key, value := range flags {
		if key&*flag != 0 {
			fmt.Printf("\t%s\n", value)
		}
	}
}

func (flag *Flag) SetFlag(setFlags ...string) {
	for _, setFlag := range setFlags {
		for key, value := range flags {
			if strings.ToLower(value) == strings.ToLower(setFlag) {
				*flag = *flag | key
				break
			}
		}
	}
}

type NTLMType1 struct {
	SSPSignature        []byte
	MessageType         uint32
	Flags               Flag
	SuppliedDomain      SecurityBuffer
	SuppliedWorkstation SecurityBuffer
	OSVersionStructure  []byte
	OtherData           []byte
}

func ReadNTLMType1(data []byte) *NTLMType1 {
	msg := NTLMType1{
		SSPSignature:        data[0:8],
		MessageType:         binary.LittleEndian.Uint32(data[8:12]),
		Flags:               (Flag)(binary.BigEndian.Uint32(data[12:16])),
		SuppliedDomain:      ReadSecurityBuffer(data[16:24]),
		SuppliedWorkstation: ReadSecurityBuffer(data[24:33]),
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
	fmt.Printf("Supplied Domain : %x\n", msg.SuppliedDomain.RawData)
	fmt.Printf("Supplied Workstation : %x\n", msg.SuppliedWorkstation.RawData)
	fmt.Printf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4]))
	fmt.Printf("\n")
}

type NTLMType2 struct {
	SSPSignature       []byte
	MessageType        uint32
	TargetName         SecurityBuffer
	Flags              Flag
	Challenge          []byte
	Context            uint32
	TargetInformation  SecurityBuffer
	OSVersionStructure []byte
	OtherData          []byte
}

type TargetInformation struct {
	Type    uint16
	Length  uint16
	Content []byte
}

func (targetInformation TargetInformation) ToBytes() []byte {
	buffer := []byte{}

	TypeBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(TypeBuffer, targetInformation.Type)
	buffer = append(buffer, TypeBuffer...)

	LengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(LengthBuffer, targetInformation.Length)
	buffer = append(buffer, LengthBuffer...)

	buffer = append(buffer, targetInformation.Content...)

	return buffer
}

// OSVersionStructure is optional and not added into it
func NewNTLMType2Short() NTLMType2 {
	msg := NTLMType2{
		SSPSignature:      append([]byte("NTLMSSP"), 0x00),
		MessageType:       uint32(0x2),
		TargetName:        NewSecurityBuffer(),
		Flags:             (Flag)(uint32(0x00)),
		Challenge:         []byte(challenge),
		Context:           uint32(0x00),
		TargetInformation: NewSecurityBuffer(),
		OtherData:         []byte{},
	}

	baseOffset := 48

	msg.TargetName.BufferLength = uint16(len(domainName))
	msg.TargetName.BufferAllocatedLength = uint16(len(domainName))
	msg.TargetName.StartOffset = uint32(baseOffset)
	msg.OtherData = append(msg.OtherData, []byte(domainName)...)

	baseOffset = baseOffset + len([]byte(domainName))

	//targetInformationDomainName := TargetInformation{
	//	Type:    uint16(0x0002),
	//	Length:  uint16(len(domainName)),
	//	Content: []byte(domainName),
	//}
	//targetInformationDomainNameBytes := targetInformationDomainName.ToBytes()

	//targetInformationServerName := TargetInformation{
	//	Type:    uint16(0x0001),
	//	Length:  uint16(len(serverName)),
	//	Content: []byte(serverName),
	//}
	//targetInformationServerNameBytes := targetInformationServerName.ToBytes()

	//targetInformationDNSDomainName := TargetInformation{
	//	Type:    uint16(0x0003),
	//	Length:  uint16(len(dnsDomainName)),
	//	Content: []byte(dnsDomainName),
	//}
	//targetInformationDNSDomainNameBytes := targetInformationDNSDomainName.ToBytes()

	//targetInformationDNSServerName := TargetInformation{
	//	Type:    uint16(0x0004),
	//	Length:  uint16(len(dnsServerName)),
	//	Content: []byte(dnsServerName),
	//}
	//targetInformationDNSServerNameBytes := targetInformationDNSServerName.ToBytes()

	//targetInformationTerminatorSubblock := TargetInformation{
	//	Type:    uint16(0x0000),
	//	Length:  uint16(0x0000),
	//	Content: []byte{},
	//}
	//targetInformationTerminatorSubblockBytes := targetInformationTerminatorSubblock.ToBytes()

	//// Subblock end
	//targetInformationBytes := []byte{}
	//targetInformationBytes = append(targetInformationBytes, targetInformationDomainNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationServerNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationDNSDomainNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationDNSServerNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationTerminatorSubblockBytes...)

	//msg.TargetInformation.BufferLength = uint16(len(targetInformationBytes))
	//msg.TargetInformation.BufferAllocatedLength = uint16(len(targetInformationBytes))
	//msg.TargetInformation.StartOffset = uint32(baseOffset)
	//msg.OtherData = append(msg.OtherData, []byte(targetInformationBytes)...)

	//baseOffset = baseOffset + len(targetInformationBytes)

	msg.Flags.SetFlag(
		//"Negotiate Unicode",
		"Negotiate OEM",
		"Request Target",
		"Negotiate NTLM",
		"Negotiate Local Call",
		"Negotiate Always Sign",
		"Target Type Domain",
		"Target Type Server",
		"Target Type Share",
		"Negotiate NTLMv2 Key",
		"Negotiate Target Info",
		"Negotiate 128",
		"Negotiate 56",
	)

	return msg
}

func (msg *NTLMType2) ToBytes() []byte {
	var result []byte

	SSPSignatureBytes := make([]byte, 8)
	copy(SSPSignatureBytes, msg.SSPSignature)
	result = append(result, SSPSignatureBytes...)

	MessageTypeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(MessageTypeBytes, msg.MessageType)
	result = append(result, MessageTypeBytes...)

	TargetNameBytes := msg.TargetName.ToBytes()
	result = append(result, TargetNameBytes...)

	FlagsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(FlagsBytes, (uint32)(msg.Flags))
	result = append(result, FlagsBytes...)

	ChallengeBytes := make([]byte, 8)
	copy(ChallengeBytes, challenge)
	result = append(result, ChallengeBytes...)

	ContextBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(ContextBytes, msg.Context)
	result = append(result, ContextBytes...)

	TargetInformationBytes := msg.TargetInformation.ToBytes()
	result = append(result, TargetInformationBytes...)

	result = append(result, msg.OtherData...)

	return result
}
