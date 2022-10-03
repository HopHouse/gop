package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

const DefaultChallenge string = "HopHouse"
const DefaultDomainName string = "smbdomain"
const serverName string = "DC"
const dnsDomainName string = "smbdomain.local"
const dnsServerName string = "dc.smbdomain.local"

type NTLMMessage interface {
	Read([]byte)
	ToString() string
	ToBytes() []byte
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

func (msg *NTLMSSP_NEGOTIATE) Read(data []byte) {
	offset := 0

	msg.SSPSignature = data[offset : offset+8]
	offset += 8

	msg.MessageType = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	msg.Flags = (Flag)(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	err := msg.SuppliedDomain.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Negotiate supplied domain : %s\n", err)
	}
	offset += 8

	err = msg.SuppliedWorkstation.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Negotiate supplied workstation : %s\n", err)
	}
	offset += 8

	msg.OSVersionStructure = data[offset : offset+8]
	offset += 8

	if len(data) > offset+8 {
		copy(msg.OtherData, data[offset:])
	}
}

func (msg *NTLMSSP_NEGOTIATE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString(fmt.Sprint("Flags :\n"))
	str.WriteString(msg.Flags.ToString())
	str.WriteString(fmt.Sprintf("Supplied Domain : %s\n", string(msg.SuppliedDomain.RawData)))
	str.WriteString(fmt.Sprintf("Supplied Workstation : %s\n", string(msg.SuppliedWorkstation.RawData)))
	str.WriteString(fmt.Sprintf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4])))
	str.WriteString(fmt.Sprintf("Other Data        : %v\n", msg.OtherData))
	str.WriteString(fmt.Sprintf("Other Data string : %v\n", string(msg.OtherData)))
	str.WriteString(fmt.Sprintf("Other Data Offset : %v\n", msg.OtherDataOffset))
	str.WriteString(fmt.Sprint("\n"))

	return str.String()
}

type NTLMSSP_CHALLENGE struct {
	SSPSignature       []byte
	MessageType        uint32
	TargetName         SecurityBuffer
	Flags              Flag
	Challenge          []byte
	Context            uint32
	TargetInformation  TargetInformation
	OSVersionStructure []byte
	OtherData          []byte
	OtherDataOffset    int
}

func (msg *NTLMSSP_CHALLENGE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString(fmt.Sprintf("TargetName      : %s\n", string(msg.TargetName.RawData)))
	str.WriteString(fmt.Sprint("Flags :\n"))
	str.WriteString(msg.Flags.ToString())
	str.WriteString(fmt.Sprintf("Challenge : %x\n", msg.Challenge))
	str.WriteString(fmt.Sprintf("Context : %d\n", msg.Context))
	str.WriteString(fmt.Sprintf("TargetInformation      :\n%s\n", msg.TargetInformation.ToString()))
	str.WriteString(fmt.Sprintf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4])))
	str.WriteString(fmt.Sprintf("Other Data        : %v\n", msg.OtherData))
	str.WriteString(fmt.Sprintf("Other Data string : %v\n", string(msg.OtherData)))
	str.WriteString(fmt.Sprintf("Other Data Offset : %v\n", msg.OtherDataOffset))
	str.WriteString(fmt.Sprint("\n"))

	return str.String()
}

// OSVersionStructure is optional and not added into it
func NewNTLMSSP_CHALLENGE(challenge string, domainName string) NTLMSSP_CHALLENGE {
	msg := NTLMSSP_CHALLENGE{
		SSPSignature:       append([]byte("NTLMSSP"), 0x00),
		MessageType:        uint32(0x2),
		TargetName:         NewEmptySecurityBuffer(),
		Flags:              (Flag)(uint32(0x00)),
		Challenge:          []byte{},
		Context:            uint32(0x00),
		TargetInformation:  TargetInformation{},
		OSVersionStructure: []byte{},
		OtherData:          []byte{},
		OtherDataOffset:    48,
	}

	msg.SetChallenge(challenge)

	domainNameBytes := []byte(domainName)
	msg.TargetName = NewSecurityBuffer(&msg.OtherData, domainNameBytes)

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

	//msg.SetSecurityBuffer(&msg.TargetInformation, []byte(targetInformationBytes))

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
	copy(ChallengeBytes, msg.Challenge)
	result = append(result, ChallengeBytes...)

	ContextBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(ContextBytes, msg.Context)
	result = append(result, ContextBytes...)

	TargetInformationBytes := msg.TargetInformation.ToBytes()
	result = append(result, TargetInformationBytes...)

	OSVersionStructureBytes := msg.OSVersionStructure
	result = append(result, OSVersionStructureBytes...)

	result = append(result, msg.OtherData...)
	return result
}

func (msg *NTLMSSP_CHALLENGE) Read(data []byte) {
	offset := 0

	msg.SSPSignature = data[offset : offset+8]
	offset += 8

	msg.MessageType = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	err := msg.TargetName.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Challenge supplied target name : %s\n", err)
	}
	offset += 8

	msg.Flags = (Flag)(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	msg.Challenge = data[offset : offset+8]
	offset += 8

	binary.LittleEndian.PutUint32(data[offset:offset+8], msg.Context)
	offset += 8

	targetInformationSecurityBuffer := NewEmptySecurityBuffer()
	err = targetInformationSecurityBuffer.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Challenge supplied target information : %s\n", err)
	}
	msg.TargetInformation.Read(targetInformationSecurityBuffer.RawData)
	offset += 8

	msg.OSVersionStructure = data[offset : offset+4]
	offset += 4

	if len(data) > offset {
		msg.OtherData = data[offset:]
	}
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
	offset := 0

	msg.SSPSignature = data[offset : offset+8]
	offset += 8

	msg.MessageType = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	msg.OtherDataOffset = 52

	err := msg.LMv2Response.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Auth LMv2 Response : %s\n", err)
	}
	offset += 8

	err = msg.NTLMv2Response.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Auth NTLMv2 Response : %s\n", err)
	}
	offset += 8

	err = msg.TargetName.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Auth target name : %s\n", err)
	}
	offset += 8

	err = msg.Username.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Auth user name : %s\n", err)
	}
	offset += 8

	err = msg.Workstation.SetSecurityBuffer(data[offset:offset+8], data)
	if err != nil {
		logger.Printf("Error while reading the NTLM Auth workstation name : %s\n", err)
	}
	offset += 8

	// Session Key is optional. If LMv2 Security Buffer has an offset equal to 52,
	// so the session key is not present
	if len(data) > 52 {
		err = msg.SessionKey.SetSecurityBuffer(data[offset:offset+8], data)
		if err != nil {
			logger.Printf("Error while reading the NTLM Auth sessoion key : %s\n", err)
		}
		offset += 8
	}

	// Flags are optional. If LMv2 Security Buffer has an offset equal to 60,
	// so the session key is present
	if len(data) > 64 {
		if msg.OtherDataOffset > 60 {
			msg.Flags = (Flag)(binary.BigEndian.Uint32(data[offset : offset+4]))
		}
		offset += 4
	}

	// Os Version is optional. If LMv2 Security Buffer has an offset equal to 64,
	// so the session key is present
	if len(data) > 64 {
		msg.OSVersionStructure = data[offset : offset+8]
		offset += 8
	}

	if len(data) > offset {
		msg.OtherData = data[offset:]
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
	str.WriteString(fmt.Sprintf("LMv2Response : %x\n", msg.LMv2Response.RawData))
	str.WriteString(fmt.Sprintf("NTLMv2Response : %x\n", msg.NTLMv2Response.RawData))
	str.WriteString(fmt.Sprintf("Targetname : %s\n", msg.TargetName.RawData))
	str.WriteString(fmt.Sprintf("UserName : %s\n", msg.Username.RawData))
	str.WriteString(fmt.Sprintf("Workstation : %s\n", msg.Workstation.RawData))
	str.WriteString(fmt.Sprintf("SessionKey : %x\n", msg.SessionKey.RawData))
	str.WriteString(fmt.Sprintf("Flags :\n"))
	str.WriteString(fmt.Sprint(msg.Flags.ToString()))
	str.WriteString(fmt.Sprintf("OS Version : %v.%v - Build %d\n", msg.OSVersionStructure[0], msg.OSVersionStructure[1], binary.LittleEndian.Uint16(msg.OSVersionStructure[2:4])))
	str.WriteString(fmt.Sprintf("\n"))

	return str.String()
}
