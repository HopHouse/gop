package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

type SecurityBuffer struct {
	BufferLength          uint16
	BufferAllocatedLength uint16
	StartOffset           uint32
	RawData               []byte
}

func NewEmptySecurityBuffer() SecurityBuffer {
	return SecurityBuffer{
		BufferLength:          uint16(0),
		BufferAllocatedLength: uint16(0),
		StartOffset:           uint32(0),
		RawData:               []byte{},
	}
}

func ReadSecurityBuffer(data []byte, start int) SecurityBuffer {
	buffer := SecurityBuffer{
		BufferLength:          binary.LittleEndian.Uint16(data[start : start+2]),
		BufferAllocatedLength: binary.LittleEndian.Uint16(data[start+2 : start+4]),
		StartOffset:           binary.LittleEndian.Uint32(data[start+4 : start+8]),
		RawData:               []byte{},
	}

	if int(buffer.BufferAllocatedLength) > 0 {
		buffer.RawData = data[int(buffer.StartOffset) : int(buffer.StartOffset)+int(buffer.BufferAllocatedLength)]
	}

	return buffer
}

func (sbuf SecurityBuffer) SetSecurityBuffer(rawData []byte, otherDataOffset int) {
	// Set the security buffer
	sbuf.BufferLength = uint16(len([]byte(rawData)))
	sbuf.BufferAllocatedLength = uint16(len([]byte(rawData)))
	sbuf.StartOffset = uint32(otherDataOffset)
	sbuf.RawData = []byte(rawData)
}

func (sbuf SecurityBuffer) ToBytes() []byte {
	buffer := make([]byte, 0)

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
	logger.Printf(sbuf.ToString())
}

func (sbuf *SecurityBuffer) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("\tBuffer length                : %d\n", sbuf.BufferLength))
	str.WriteString(fmt.Sprintf("\tBuffer Allocated length      : %d\n", sbuf.BufferAllocatedLength))
	str.WriteString(fmt.Sprintf("\tOffset                       : %d\n", sbuf.StartOffset))
	str.WriteString(fmt.Sprintf("\tData Bytes                   : %b\n", sbuf.RawData))
	str.WriteString(fmt.Sprintf("\tData String                  : %s\n", ByteSliceToString(sbuf.RawData[:])))

	return str.String()
}
