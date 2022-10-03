package ntlm

import (
	"bytes"
	"encoding/binary"
)

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

func (targetInformation TargetInformation) Read(buffer []byte) {
	targetInformation.Type = binary.LittleEndian.Uint16(buffer[0:2])
	targetInformation.Length = binary.LittleEndian.Uint16(buffer[2:4])
	targetInformation.Content = bytes.Runes(buffer[4:targetInformation.Length])
}
