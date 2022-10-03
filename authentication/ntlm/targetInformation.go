package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type TargetInformationBlock struct {
	Type    uint16
	Length  uint16
	Content []byte
}
type TargetInformation struct {
	Blocks []TargetInformationBlock
}

const (
	TargetInformationServerName            = uint16(1)
	TargetInformationDomainName            = uint16(2)
	TargetInformationFullyQualifiedDNSName = uint16(3)
	TargetInformationDNSDomainName         = uint16(4)
	TargetInformationParentDNSDomainName   = uint16(5)
)

func (t *TargetInformationBlock) ToBytes() []byte {
	buffer := []byte{}

	TypeBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(TypeBuffer, t.Type)
	buffer = append(buffer, TypeBuffer...)

	LengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(LengthBuffer, t.Length)
	buffer = append(buffer, LengthBuffer...)

	buffer = append(buffer, []byte(string(t.Content))...)

	return buffer
}

func (t *TargetInformationBlock) Read(buffer []byte) {
	t.Type = binary.LittleEndian.Uint16(buffer[0:2])
	t.Length = binary.LittleEndian.Uint16(buffer[2:4])
	t.Content = buffer[4 : 4+t.Length]
}

func (t *TargetInformationBlock) ToString() string {
	var str strings.Builder

	typeString := ""

	switch t.Type {
	case TargetInformationServerName:
		typeString = "Server name"
	case TargetInformationDomainName:
		typeString = "Domain name"
	case TargetInformationFullyQualifiedDNSName:
		typeString = "Fully-qualified DNS host name"
	case TargetInformationDNSDomainName:
		typeString = "DNS domain name"
	case TargetInformationParentDNSDomainName:
		typeString = "Parent DNS domain name"
	default:
		typeString = fmt.Sprintf("Unknown type %d", t.Type)
	}

	str.WriteString(fmt.Sprintf("\tType                   : %s\n", typeString))
	str.WriteString(fmt.Sprintf("\tLength                 : %d\n", t.Length))
	str.WriteString(fmt.Sprintf("\tContent                : %s\n", string(t.Content)))

	return str.String()
}

func (t *TargetInformation) ToBytes() []byte {
	buffer := []byte{}

	for _, block := range t.Blocks {
		buffer = append(buffer, block.ToBytes()...)
	}
	return buffer
}

func (t *TargetInformation) Read(buffer []byte) {
	for offset := 0; offset < len(buffer); {
		block := TargetInformationBlock{}
		block.Read(buffer[offset:])

		// block with type 0 ends the TargetInformation blocks
		t.Blocks = append(t.Blocks, block)

		headerSize := 4
		offset += headerSize + int(block.Length)
	}
}

func (t *TargetInformation) ToString() string {
	var str strings.Builder
	for _, block := range t.Blocks {
		str.WriteString(block.ToString())
	}
	return str.String()
}
