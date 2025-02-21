package ntlm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

type SecurityBuffer struct {
	BufferLength    uint16
	BufferMaxLength uint16
	StartOffset     uint32
	Payload         []byte
}

func NewEmptySecurityBuffer() SecurityBuffer {
	return SecurityBuffer{
		BufferLength:    uint16(0),
		BufferMaxLength: uint16(0),
		StartOffset:     uint32(0),
		Payload:         []byte{},
	}
}

func NewSecurityBufferForData(data []byte, offset uint32) SecurityBuffer {
	return SecurityBuffer{
		BufferLength:    uint16(len(data)),
		BufferMaxLength: uint16(len(data)),
		StartOffset:     offset,
		Payload:         data,
	}
}

func ReadSecurityBuffer(data []byte, start int) SecurityBuffer {
	buffer := SecurityBuffer{}

	buffer.BufferLength = binary.LittleEndian.Uint16(data[start : start+2])
	buffer.BufferMaxLength = binary.LittleEndian.Uint16(data[start+2 : start+4])
	buffer.StartOffset = binary.LittleEndian.Uint32(data[start+4 : start+8])
	buffer.Payload = make([]byte, buffer.BufferLength)

	copy(buffer.Payload, data[int(buffer.StartOffset):int(buffer.StartOffset)+int(buffer.BufferLength)])

	return buffer
}

func (sbuf SecurityBuffer) SetSecurityBuffer(data []byte, start int) {
	sbuf.BufferLength = binary.LittleEndian.Uint16(data[start : start+2])
	sbuf.BufferMaxLength = binary.LittleEndian.Uint16(data[start+2 : start+4])
	sbuf.StartOffset = binary.LittleEndian.Uint32(data[start+4 : start+8])
	sbuf.Payload = make([]byte, sbuf.BufferLength)

	copy(sbuf.Payload, data[int(sbuf.StartOffset):int(sbuf.StartOffset)+int(sbuf.BufferLength)])
}

func (sbuf SecurityBuffer) ToBytes() ([]byte, []byte) {
	var data bytes.Buffer
	var payload bytes.Buffer

	binary.Write(&data, binary.LittleEndian, sbuf.BufferLength)
	binary.Write(&data, binary.LittleEndian, sbuf.BufferMaxLength)
	binary.Write(&data, binary.LittleEndian, sbuf.StartOffset)

	binary.Write(&payload, binary.LittleEndian, sbuf.Payload[:])

	return data.Bytes(), payload.Bytes()
}

func (sbuf *SecurityBuffer) PrintSecurityBuffer() {
	logger.Printf(sbuf.ToString())
}

func (sbuf *SecurityBuffer) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("\tBuffer length                : %d\n", sbuf.BufferLength))
	str.WriteString(fmt.Sprintf("\tBuffer Allocated length      : %d\n", sbuf.BufferMaxLength))
	str.WriteString(fmt.Sprintf("\tOffset                       : %d\n", sbuf.StartOffset))
	str.WriteString(fmt.Sprintf("\tData Bytes                   : %b\n", sbuf.Payload))
	str.WriteString(fmt.Sprintf("\tData String                  : %s\n", ByteSliceToString(sbuf.Payload[:])))

	return str.String()
}
