package ntlm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

const (
	domainName string = "WORKGROUP"
)

var Challenge string = string([]byte{0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41})

// const serverName string = "DC"
// const dnsDomainName string = "smbdomain.local"
// const dnsServerName string = "dc.smbdomain.local"

type NTLMMessage interface {
	Read([]byte) error
	ToString() string
	ToBytes() []byte
}

// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/b34032e5-3aae-4bc6-84c3-c6d80eadf7f2
type NTLMSSP_NEGOTIATE struct {
	SSPSignature        uint64
	MessageType         uint32
	Flags               uint32
	SuppliedDomain      SecurityBuffer
	SuppliedWorkstation SecurityBuffer
	OSVersionStructure  NTLMSSP_VERSION
}

func (msg *NTLMSSP_NEGOTIATE) ToBytes() []byte {
	var data bytes.Buffer

	binary.Write(&data, binary.LittleEndian, msg.SSPSignature)
	binary.Write(&data, binary.LittleEndian, msg.MessageType)
	binary.Write(&data, binary.LittleEndian, msg.Flags)

	suppliedDomainSecurityBuffer, suppliedDomainPayload := msg.SuppliedDomain.ToBytes()
	binary.Write(&data, binary.LittleEndian, suppliedDomainSecurityBuffer)

	suppliedWorkstationSecurityBuffer, suppliedWorkstationPayload := msg.SuppliedWorkstation.ToBytes()
	binary.Write(&data, binary.LittleEndian, suppliedWorkstationSecurityBuffer)

	binary.Write(&data, binary.LittleEndian, msg.OSVersionStructure)

	binary.Write(&data, binary.LittleEndian, suppliedDomainPayload)
	binary.Write(&data, binary.LittleEndian, suppliedWorkstationPayload)

	return data.Bytes()
}

func (msg *NTLMSSP_NEGOTIATE) Read(data []byte) error {
	if slices.Compare(data[0:8], []byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0x00}) != 0 {
		err := fmt.Errorf("ntlmssp negotiate signature does not begin with 'N', 'T', 'L', 'M', 'S', 'S', 'P', '\\0' : %x - %s", data[0:8], data[0:8])
		return err
	}
	msg.SSPSignature = binary.LittleEndian.Uint64(data[0:8])

	msg.MessageType = binary.LittleEndian.Uint32(data[8:12])
	if msg.MessageType != 1 {
		err := fmt.Errorf("ntlmssp negotiate message type is different from 0x00000001 : %x", msg.MessageType)
		return err
	}

	msg.Flags = binary.LittleEndian.Uint32(data[12:16])

	if msg.Flags&uint32(NTLMSSP_NEGOTIATE_DOMAIN_SUPPLIED) != 0 {
		// A domain name is supplied to the buffer
		msg.SuppliedDomain = ReadSecurityBuffer(data, 16)
	} else {
		msg.SuppliedDomain = SecurityBuffer{
			BufferLength:    0,
			BufferMaxLength: 0,
			// DomainNameBufferOffset field SHOULD be set to the offset from the beginning of the NEGOTIATE_MESSAGE to where the DomainName would be in Payload if it were present.
			//
			//  NTTLMSSP_NEGOTIATE STRUCT :
			// 		SSPSignature        	uint64				// 64 bits = 8 bytes
			// 		MessageType         	uint32				// 32 bits = 4 bytes
			// 		Flags               	uint32				// 32 bits = 4 bytes
			// 		SuppliedDomain         	SecurityBuffer  	// 64 byts = 8 bytes
			//  	SuppliedWorkstation    	SecurityBuffer  	// 64 byts = 8 bytes
			//  	OSVersionStructure     	uint64          	// 64 byts = 8 bytes
			//
			StartOffset: 0,
			Payload:     []byte{},
		}
	}

	if msg.Flags&uint32(NTLMSSP_NEGOTIATE_WORKSTATION_SUPPLIED) != 0 {
		// A workstation name is supplied to the buffer
		msg.SuppliedWorkstation = ReadSecurityBuffer(data, 24)
	} else {
		msg.SuppliedWorkstation = SecurityBuffer{
			BufferLength:    0,
			BufferMaxLength: 0,
			// DomainNameBufferOffset field SHOULD be set to the offset from the beginning of the NEGOTIATE_MESSAGE to where the DomainName would be in Payload if it were present.
			//
			//  NTTLMSSP_NEGOTIATE STRUCT :
			// 		SSPSignature        	uint64				// 64 bits = 8 bytes
			// 		MessageType         	uint32				// 32 bits = 4 bytes
			// 		Flags               	uint32				// 32 bits = 4 bytes
			// 		SuppliedDomain         	SecurityBuffer  	// 64 byts = 8 bytes
			//  	SuppliedWorkstation    	SecurityBuffer  	// 64 byts = 8 bytes
			//  	OSVersionStructure     	uint64          	// 64 byts = 8 bytes
			//
			StartOffset: 0,
			Payload:     []byte{},
		}
	}

	msg.OSVersionStructure.Read(data[32:40])

	return nil
}

func (msg *NTLMSSP_NEGOTIATE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %x\n", msg.SSPSignature))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString("Flags 								:\n")
	str.WriteString((*Flag)(&msg.Flags).ToString())
	str.WriteString(fmt.Sprintf("Supplied Domain 		: 0x%x\n", msg.SuppliedDomain.Payload))
	str.WriteString(fmt.Sprintf("Supplied Workstation	: 0x%x\n", msg.SuppliedWorkstation.Payload))
	str.WriteString(fmt.Sprintf("OS Version 			:\n%s", msg.OSVersionStructure.ToString()))

	return str.String()
}

type NTLMSSP_CHALLENGE struct {
	SSPSignature       uint64
	MessageType        uint32
	TargetName         SecurityBuffer
	Flags              uint32
	Challenge          uint64
	Reserved           uint64
	TargetInformation  SecurityBuffer
	OSVersionStructure NTLMSSP_VERSION
}

func (msg *NTLMSSP_CHALLENGE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %x\n", msg.SSPSignature))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString(fmt.Sprintf("TargetName             :\n%s\n", msg.TargetName.ToString()))

	str.WriteString(fmt.Sprint("Flags 					:\n"))
	str.WriteString((*Flag)(&msg.Flags).ToString())

	str.WriteString(fmt.Sprintf("Challenge 				: 0x%x\n", msg.Challenge))
	str.WriteString(fmt.Sprintf("Reserved  				: %d\n", msg.Reserved))
	str.WriteString(fmt.Sprintf("TargetInformation      :\n%s\n", msg.TargetInformation.ToString()))
	str.WriteString(fmt.Sprintf("OS Version 			:\n%s\n", msg.OSVersionStructure.ToString()))

	return str.String()
}

// OSVersionStructure is optional and not added into it
func NewNTLMSSP_CHALLENGEShort() NTLMSSP_CHALLENGE {
	msg := NTLMSSP_CHALLENGE{}

	msg.SSPSignature = binary.LittleEndian.Uint64([]byte("NTLMSSP\x00"))
	msg.MessageType = 0x02
	msg.TargetName = NewEmptySecurityBuffer()
	msg.Flags = 0x00
	msg.Challenge = 0x00
	msg.Reserved = 0x00
	msg.TargetInformation = NewEmptySecurityBuffer()
	msg.OSVersionStructure = NTLMSSP_VERSION{
		ProductMajorVersion: 10,
		ProductMinorVersion: 0,
		ProductBuild:        26100,
		NTLMRevisionCurrent: 15,
	}

	(*Flag)(&msg.Flags).SetFlag(
		// 	// "Negotiate Local Call",
		// 	// "Negotiate Always Sign",
		// 	// "Target Type Domain",
		// 	// "Target Type Share",
		"Negotiate 56",
		"Negotiate Key Exchange",
		"Negotiate 128",
		"Negotiate Version",
		"Negotiate Target Info",
		"Negotiate NTLMv2 Key",
		"Target Type Server",
		"Negotiate NTLM",
		"Negotiate Sign",
		"Request Target",
		"Negotiate OEM",
		"Negotiate Unicode",
	)
	// logger.Println((*Flag)(&msg.Flags).ToString())

	msg.SetChallenge(Challenge)

	// 	NTLMSSP_CHALLENGE struct :
	// 		SSPSignature       		uint64 				// 64 bits = 8 bytes
	// 		MessageType        		uint32				// 32 bits = 4 bytes
	// 		TargetName         		SecurityBuffer		// 64 bits = 8 bytes
	// 		Flags              		uint32				// 32 bits = 4 bytes
	// 		Challenge          		uint64				// 64 bits = 8 bytes
	// 		Reserved           		uint64				// 64 bits = 8 bytes
	// 		TargetInformation  		SecurityBuffer 		// 64 bits = 8 bytes
	// 		OSVersionStructure 		NTLMSSP_VERSION		// 64 bits = 8 bytes
	// 		OtherData          		[]byte
	// 		OtherDataOffset    		int
	//
	// So the startOffset must be set to 8 + 4 + 8 + 4 + 8 + 8 + 8  = 48 + 8 if version flag is set .
	payloadOffset := uint32(48)
	if msg.Flags&uint32(NTLMSSP_NEGOTIATE_VERSION) != 0 {
		payloadOffset = 56
	}

	// TODO : Fix it here. The domainNameByte must be in unicode to bytes, so uint16 instead of uint8
	// domainNameByte := []byte(domainName)
	domainNameByte := []byte{0x57, 0x0, 0x4f, 0x0, 0x52, 0x0, 0x4b, 0x0, 0x47, 0x0, 0x52, 0x0, 0x4f, 0x0, 0x55, 0x0, 0x50, 0x0}
	msg.TargetName = NewSecurityBufferForData(domainNameByte, payloadOffset)

	// payload become offet + msg.TargetName.Length
	payloadOffset += uint32(msg.TargetName.BufferLength)

	// msg.TargetInformation = NewSecurityBufferForData([]byte{0x00, 0x00, 0x00, 0x00}, payloadOffset)
	msg.TargetInformation = NewSecurityBufferForData([]byte{0x1, 0x0, 0x16, 0x0, 0x73, 0x0, 0x65, 0x0, 0x72, 0x0, 0x76, 0x0, 0x65, 0x0, 0x72, 0x0, 0x5f, 0x0, 0x6e, 0x0, 0x61, 0x0, 0x6d, 0x0, 0x65, 0x0, 0x3, 0x0, 0x16, 0x0, 0x73, 0x0, 0x65, 0x0, 0x72, 0x0, 0x76, 0x0, 0x65, 0x0, 0x72, 0x0, 0x5f, 0x0, 0x6e, 0x0, 0x61, 0x0, 0x6d, 0x0, 0x65, 0x0, 0x2, 0x0, 0x12, 0x0, 0x57, 0x0, 0x4f, 0x0, 0x52, 0x0, 0x4b, 0x0, 0x47, 0x0, 0x52, 0x0, 0x4f, 0x0, 0x55, 0x0, 0x50, 0x0, 0x4, 0x0, 0x12, 0x0, 0x57, 0x0, 0x4f, 0x0, 0x52, 0x0, 0x4b, 0x0, 0x47, 0x0, 0x52, 0x0, 0x4f, 0x0, 0x55, 0x0, 0x50, 0x0, 0x7, 0x0, 0x8, 0x0, 0x80, 0x5f, 0xfe, 0x65, 0x6e, 0x84, 0xdb, 0x1, 0x0, 0x0, 0x0, 0x0}, payloadOffset)
	//
	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/801a4681-8809-4be9-ab0d-61dcfe762786
	//
	// TargetInfo (variable): If TargetInfoLen does not equal 0x0000, TargetInfo MUST be a byte array that contains a sequence of AV_PAIR structures. The AV_PAIR structure is defined in section 2.2.2.1. The length of each AV_PAIR is determined by its AvLen field (plus 4 bytes).
	//
	// Note An AV_PAIR structure can start on any byte alignment and the sequence of AV_PAIRs has no padding between structures.
	//
	// The sequence MUST be terminated by an AV_PAIR structure with an AvId field of MsvAvEOL. The total length of the TargetInfo byte array is the sum of the lengths, in bytes, of the AV_PAIR structures it contains.
	//

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

	msg.Challenge = binary.LittleEndian.Uint64([]byte(challenge))
}

func (msg *NTLMSSP_CHALLENGE) ToBytes() []byte {
	var data bytes.Buffer
	var payload bytes.Buffer

	binary.Write(&data, binary.LittleEndian, msg.SSPSignature)
	binary.Write(&data, binary.LittleEndian, msg.MessageType)

	targetNameSecurityBuffer, targetNamePayload := msg.TargetName.ToBytes()
	binary.Write(&data, binary.LittleEndian, targetNameSecurityBuffer)
	binary.Write(&payload, binary.LittleEndian, targetNamePayload)

	binary.Write(&data, binary.LittleEndian, msg.Flags)
	binary.Write(&data, binary.LittleEndian, msg.Challenge)
	binary.Write(&data, binary.LittleEndian, msg.Reserved)

	targetInformationSecurityBuffer, targetInformationPayload := msg.TargetInformation.ToBytes()
	binary.Write(&data, binary.LittleEndian, targetInformationSecurityBuffer)
	binary.Write(&payload, binary.LittleEndian, targetInformationPayload)

	if msg.Flags&uint32(NTLMSSP_NEGOTIATE_VERSION) != 0 {
		binary.Write(&data, binary.LittleEndian, msg.OSVersionStructure.ToBytes())
	}

	// Write payload at the end
	data.Write(targetNamePayload)
	// for i := 0; i < len(targetNamePayload)%8; i++ {
	// 	data.WriteByte(0x00)
	// }

	data.Write(targetInformationPayload)
	// for i := 0; i < len(targetInformationPayload)%8; i++ {
	// 	data.WriteByte(0x00)
	// }

	return data.Bytes()
}

func (msg *NTLMSSP_CHALLENGE) Read(data []byte) error {
	msg.SSPSignature = binary.LittleEndian.Uint64(data[0:8])
	msg.MessageType = binary.LittleEndian.Uint32(data[8:12])
	msg.TargetName = ReadSecurityBuffer(data, 12)
	msg.Flags = binary.LittleEndian.Uint32(data[20:24])
	msg.Challenge = binary.LittleEndian.Uint64(data[24:32])
	msg.Reserved = binary.LittleEndian.Uint64(data[32:40])
	msg.TargetInformation = ReadSecurityBuffer(data, 40)

	if msg.Flags&uint32(NTLMSSP_NEGOTIATE_VERSION) != 0 {
		err := msg.OSVersionStructure.Read(data[48:56])
		if err != nil {
			return err
		}
	}

	return nil
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
	OSVersionStructure NTLMSSP_VERSION
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
		msg.OSVersionStructure.Read(data[64:72])
	}
	msg.OtherData = data[72:]
}

func (msg *NTLMSSP_AUTH) ToBytes() []byte {
	logger.Printf("[!] MTLMSSP_AUTH.ToBytes() Not implemented")
	return []byte{}
}

func (msg *NTLMSSP_AUTH) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))

	str.WriteString(fmt.Sprintf("NTLMSSP Signature      : %s\n", string(msg.SSPSignature)))
	str.WriteString(fmt.Sprintf("NTLM Message Type      : %v\n", msg.MessageType))
	str.WriteString(fmt.Sprintf("LMv2Response : 0x%x\n", msg.LMv2Response.Payload))
	str.WriteString(fmt.Sprintf("NTLMv2Response : 0x%x\n", msg.NTLMv2Response.Payload))
	str.WriteString(fmt.Sprintf("Targetname : %s\n", msg.TargetName.Payload))
	str.WriteString(fmt.Sprintf("UserName : %s\n", msg.Username.Payload))
	str.WriteString(fmt.Sprintf("Workstation : %s\n", msg.Workstation.Payload))
	str.WriteString(fmt.Sprintf("SessionKey : 0x%x\n", msg.SessionKey.Payload))
	str.WriteString("Flags : \n")
	str.WriteString(msg.Flags.ToString())
	str.WriteString(fmt.Sprintf("OS Version :\n%s\n", msg.OSVersionStructure.ToString()))

	return str.String()
}

func (msg *NTLMSSP_AUTH) SetSecurityBuffer(sbuf *SecurityBuffer, rawData []byte) {
	// Set the security buffer
	sbuf.SetSecurityBuffer(rawData, msg.OtherDataOffset)

	// Add data to OtherData
	msg.OtherData = append(msg.OtherData, []byte(rawData)...)

	msg.OtherDataOffset = msg.OtherDataOffset + len([]byte(rawData))
}

type NTLMSSP_VERSION struct {
	ProductMajorVersion uint8
	ProductMinorVersion uint8
	ProductBuild        uint16
	Reserved            [3]byte
	NTLMRevisionCurrent uint8
}

func (n *NTLMSSP_VERSION) Read(data []byte) error {
	n.ProductMajorVersion = data[0]
	n.ProductMinorVersion = data[1]
	n.ProductBuild = binary.LittleEndian.Uint16(data[2:4])
	n.Reserved = [3]byte(data[4:7])
	n.NTLMRevisionCurrent = data[7]

	return nil
}

func (n *NTLMSSP_VERSION) ToBytes() []byte {
	var data bytes.Buffer

	binary.Write(&data, binary.LittleEndian, n.ProductMajorVersion)
	binary.Write(&data, binary.LittleEndian, n.ProductMinorVersion)
	binary.Write(&data, binary.LittleEndian, n.ProductBuild)
	binary.Write(&data, binary.LittleEndian, n.Reserved)
	binary.Write(&data, binary.LittleEndian, n.NTLMRevisionCurrent)

	return data.Bytes()
}

func (n *NTLMSSP_VERSION) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("Product Major Version      : %d\n", n.ProductMajorVersion))
	str.WriteString(fmt.Sprintf("Product Minor Version      : %d\n", n.ProductMinorVersion))
	str.WriteString(fmt.Sprintf("Product Build              : %d\n", n.ProductBuild))
	str.WriteString(fmt.Sprintf("Product Reserved           : %x\n", n.Reserved))
	str.WriteString(fmt.Sprintf("NTLM Revision Current      : %d\n", n.NTLMRevisionCurrent))

	return str.String()
}
