package ntlm

import (
	"encoding/binary"

	"github.com/hophouse/gop/utils/logger"
)

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
	logger.Printf("\tBuffer length                : %x\n", sbuf.BufferLength)
	logger.Printf("\tBuffer Allocated length      : %x\n", sbuf.BufferAllocatedLength)
	logger.Printf("\tOffset                       : %x\n", sbuf.StartOffset)
}
