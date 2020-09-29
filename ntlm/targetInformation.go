package ntlm

import "encoding/binary"

type TargetInformation struct {
	Type    uint16
	Length  uint16
	Content []rune
}

func (targetInformation TargetInformation) ToBytes() []byte {
	buffer := []byte{}

	TypeBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(TypeBuffer, targetInformation.Type)
	buffer = append(buffer, TypeBuffer...)

	LengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(LengthBuffer, targetInformation.Length)
	buffer = append(buffer, LengthBuffer...)

	buffer = append(buffer, []byte(string(targetInformation.Content))...)

	return buffer
}
