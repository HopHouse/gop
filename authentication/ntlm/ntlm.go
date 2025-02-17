package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

const (
	Challenge  string = "HopHouse"
	domainName string = "smbdomain"
)

// const serverName string = "DC"
// const dnsDomainName string = "smbdomain.local"
// const dnsServerName string = "dc.smbdomain.local"

type NTLMMessage interface {
	Read([]byte)
	ToString() string
	ToBytes() []byte
	SetSecurityBuffer(sbuf *SecurityBuffer)
}

type NTLMSSP_NEGOTIATE struct {
	SSPSignature        []byte
	MessageType         uint32
	Flags               Flag
	SuppliedDomain      SecurityBuffer
	SuppliedWorkstation SecurityBuffer
	OSVersionStructure  []byte
	OtherData           []byte
	OtherDataOffset     int
}

func (msg *NTLMSSP_NEGOTIATE) SetSecurityBuffer(sbuf *SecurityBuffer, rawData []byte) {
	// Set the security buffer
	sbuf.SetSecurityBuffer(rawData, msg.OtherDataOffset)

	// Add data to OtherData
	msg.OtherData = append(msg.OtherData, []byte(rawData)...)

	msg.OtherDataOffset = msg.OtherDataOffset + len([]byte(rawData))
}

func (msg *NTLMSSP_NEGOTIATE) Read(data []byte) {
	msg.SSPSignature = data[0:8]
	msg.MessageType = binary.LittleEndian.Uint32(data[8:12])
	msg.Flags = (Flag)(binary.BigEndian.Uint32(data[12:16]))
	msg.SuppliedDomain = ReadSecurityBuffer(data, 16)
	msg.SuppliedWorkstation = ReadSecurityBuffer(data, 24)
	msg.OSVersionStructure = data[32:41]
	msg.OtherData = data[40:]
}

func (msg *NTLMSSP_NEGOTIATE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString("Flags : \n")
	str.WriteString(msg.Flags.ToString())
	str.WriteString(fmt.Sprintf("Supplied Domain : %x\n", msg.SuppliedDomain.RawData))
	str.WriteString(fmt.Sprintf("Supplied Workstation : %x\n", msg.SuppliedWorkstation.RawData))
	str.WriteString(fmt.Sprintf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4])))
	str.WriteString("\n")

	return str.String()
}

type NTLMSSP_CHALLENGE struct {
	SSPSignature       []byte
	MessageType        uint32
	TargetName         SecurityBuffer
	Flags              Flag
	Challenge          []byte
	Context            uint32
	TargetInformation  SecurityBuffer
	OSVersionStructure []byte
	OtherData          []byte
	OtherDataOffset    int
}

func (msg *NTLMSSP_CHALLENGE) SetSecurityBuffer(sbuf *SecurityBuffer, rawData []byte) {

	// Set the security buffer
	sbuf.SetSecurityBuffer(rawData, msg.OtherDataOffset)

	// Add data to OtherData
	msg.OtherData = append(msg.OtherData, []byte(rawData)...)

	msg.OtherDataOffset = msg.OtherDataOffset + len([]byte(rawData))
}

// OSVersionStructure is optional and not added into it
func NewNTLMSSP_CHALLENGEShort() NTLMSSP_CHALLENGE {
	msg := NTLMSSP_CHALLENGE{
		SSPSignature:      append([]byte("NTLMSSP"), 0x00),
		MessageType:       uint32(0x2),
		TargetName:        NewSecurityBuffer(),
		Flags:             (Flag)(uint32(0x00)),
		Challenge:         []byte(Challenge),
		Context:           uint32(0x00),
		TargetInformation: NewSecurityBuffer(),
		OtherData:         []byte{},
		OtherDataOffset:   48,
	}

	msg.SetSecurityBuffer(&msg.TargetName, []byte(domainName))

	//targetInformationDomainNameBytes := TargetInformation{
	//	Type:    uint16(0x0002),
	//	Length:  uint16(len(domainName)),
	//	Content: []rune(domainName),
	//}.ToBytes()

	//targetInformationServerNameBytes := TargetInformation{
	//	Type:    uint16(0x0001),
	//	Length:  uint16(len(serverName)),
	//	Content: []rune(serverName),
	//}.ToBytes()

	//targetInformationDNSDomainNameBytes := TargetInformation{
	//	Type:    uint16(0x0003),
	//	Length:  uint16(len(dnsDomainName)),
	//	Content: []rune(dnsDomainName),
	//}.ToBytes()

	//targetInformationDNSServerNameBytes := TargetInformation{
	//	Type:    uint16(0x0004),
	//	Length:  uint16(len(dnsServerName)),
	//	Content: []rune(dnsServerName),
	//}.ToBytes()

	//targetInformationTerminatorSubblockBytes := TargetInformation{
	//	Type:    uint16(0x0000),
	//	Length:  uint16(0x0000),
	//	Content: []rune(""),
	//}.ToBytes()

	//// Subblock end
	//targetInformationBytes := []byte{}
	//targetInformationBytes = append(targetInformationBytes, targetInformationDomainNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationServerNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationDNSDomainNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationDNSServerNameBytes...)
	//targetInformationBytes = append(targetInformationBytes, targetInformationTerminatorSubblockBytes...)

	// msg.SetSecurityBuffer(&msg.TargetInformation, []byte(targetInformationBytes))

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

func (msg *NTLMSSP_CHALLENGE) SetChallenge(challenge string) {
	if len(challenge) > 8 {
		logger.Printf("[!] Provided challege %s is too long. Expected 8 bytes. It will be truncated. Challenge will be : %s\n", challenge, challenge[0:7])
		challenge = challenge[0:7]
	}
	if len(challenge) < 8 {
		logger.Printf("[!] Provided challege %s is too small. Expected 8 bytes. It will be padded with \"0\". ", challenge)
		for i := len(challenge); i <= 8; i++ {
			challenge = challenge + "0"
		}
		logger.Printf("Challenge wille be : %s\n", challenge)
		challenge = challenge[0:7]
	}

	msg.Challenge = []byte(challenge)
}

func (msg *NTLMSSP_CHALLENGE) ToBytes() []byte {
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
	copy(ChallengeBytes, Challenge)
	result = append(result, ChallengeBytes...)

	ContextBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(ContextBytes, msg.Context)
	result = append(result, ContextBytes...)

	TargetInformationBytes := msg.TargetInformation.ToBytes()
	result = append(result, TargetInformationBytes...)

	result = append(result, msg.OtherData...)
	return result
}

type NTLMSSP_AUTH struct {
	// 				Description 						Content
	// 0			NTLMSSP Signature 					Null-terminated ASCII "NTLMSSP" (0x4e544c4d53535000)
	// 8			NTLM Message Type 					long (0x03000000)
	// 12			LM/LMv2 Response 					security buffer
	// 20			NTLM/NTLMv2 Response 				security buffer
	// 28			Target Name 						security buffer
	// 36			User Name 							security buffer
	// 44			Workstation Name 					security buffer
	// (52)			Session Key (optional) 				security buffer
	// (60)			Flags (optional) 					long
	// (64)			OS Version Structure (Optional) 	8 bytes
	// 52 (64) (72) Start of data block
	SSPSignature       []byte
	MessageType        uint32
	LMv2Response       SecurityBuffer
	NTLMv2Response     SecurityBuffer
	TargetName         SecurityBuffer
	Username           SecurityBuffer
	Workstation        SecurityBuffer
	SessionKey         SecurityBuffer
	Flags              Flag
	OSVersionStructure []byte
	OtherData          []byte
	OtherDataOffset    int
}

func (msg *NTLMSSP_AUTH) Read(data []byte) {
	msg.SSPSignature = data[0:8]
	msg.MessageType = binary.LittleEndian.Uint32(data[8:12])

	msg.OtherDataOffset = 52

	msg.LMv2Response = ReadSecurityBuffer(data, 12)
	if msg.OtherDataOffset < int(msg.LMv2Response.StartOffset) {
		msg.OtherDataOffset = int(msg.LMv2Response.StartOffset)
	}

	msg.NTLMv2Response = ReadSecurityBuffer(data, 20)
	if msg.OtherDataOffset < int(msg.NTLMv2Response.StartOffset) {
		msg.OtherDataOffset = int(msg.NTLMv2Response.StartOffset)
	}

	msg.TargetName = ReadSecurityBuffer(data, 28)
	if msg.OtherDataOffset < int(msg.TargetName.StartOffset) {
		msg.OtherDataOffset = int(msg.TargetName.StartOffset)
	}

	msg.Username = ReadSecurityBuffer(data, 36)
	if msg.OtherDataOffset < int(msg.Username.StartOffset) {
		msg.OtherDataOffset = int(msg.Username.StartOffset)
	}

	msg.Workstation = ReadSecurityBuffer(data, 44)
	if msg.OtherDataOffset < int(msg.Workstation.StartOffset) {
		msg.OtherDataOffset = int(msg.Workstation.StartOffset)
	}

	// Session Key is optional. If LMv2 Security Buffer has an offset equal to 52,
	// so the session key is not present
	if msg.OtherDataOffset > 52 {
		msg.SessionKey = ReadSecurityBuffer(data, 52)
	}

	// Flags are optional. If LMv2 Security Buffer has an offset equal to 60,
	// so the session key is present
	if msg.OtherDataOffset > 60 {
		msg.Flags = (Flag)(binary.BigEndian.Uint32(data[60:64]))
	}

	// Os Version is optional. If LMv2 Security Buffer has an offset equal to 64,
	// so the session key is present
	if msg.OtherDataOffset > 64 {
		msg.OSVersionStructure = data[64:72]
	}
	msg.OtherData = data[72:]
}

func (msg *NTLMSSP_AUTH) ToBytes() []byte {
	return []byte{}
}

func (msg *NTLMSSP_AUTH) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString(fmt.Sprintf("LMv2Response : %x\n", msg.LMv2Response.RawData))
	str.WriteString(fmt.Sprintf("NTLMv2Response : %x\n", msg.NTLMv2Response.RawData))
	str.WriteString(fmt.Sprintf("Targetname : %s\n", msg.TargetName.RawData))
	str.WriteString(fmt.Sprintf("UserName : %s\n", msg.Username.RawData))
	str.WriteString(fmt.Sprintf("Workstation : %s\n", msg.Workstation.RawData))
	str.WriteString(fmt.Sprintf("SessionKey : %x\n", msg.SessionKey.RawData))
	str.WriteString("Flags : \n")
	str.WriteString(msg.Flags.ToString())
	str.WriteString(fmt.Sprintf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4])))
	str.WriteString("\n")

	return str.String()
}

func (msg *NTLMSSP_AUTH) SetSecurityBuffer(sbuf *SecurityBuffer, rawData []byte) {
	// Set the security buffer
	sbuf.SetSecurityBuffer(rawData, msg.OtherDataOffset)

	// Add data to OtherData
	msg.OtherData = append(msg.OtherData, []byte(rawData)...)

	msg.OtherDataOffset = msg.OtherDataOffset + len([]byte(rawData))
}
