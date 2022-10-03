package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/hophouse/gop/utils/logger"
	"golang.org/x/sys/windows"
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
	buffer := SecurityBuffer{}

	buffer.SetSecurityBuffer(data[start:], data)

	return buffer
}

func (sbuf *SecurityBuffer) SetSecurityBuffer(securityBufferHeader []byte, rawData []byte) error {
	// Set the security buffer
	sbuf.BufferLength = binary.LittleEndian.Uint16(securityBufferHeader[0:2])

	sbuf.BufferAllocatedLength = binary.LittleEndian.Uint16(securityBufferHeader[2:4])

	sbuf.StartOffset = binary.LittleEndian.Uint32(securityBufferHeader[4:8])

	sbuf.RawData = make([]byte, sbuf.BufferAllocatedLength)

	n := copy(sbuf.RawData, rawData[sbuf.StartOffset:int(sbuf.StartOffset)+int(sbuf.BufferLength)])

	if n != int(sbuf.BufferLength) {
		err := fmt.Errorf("copied %d data, but expected %d to be copied", n, sbuf.BufferLength)
		return err
	}

	return nil
}

func NewSecurityBuffer(ntlmRawData *[]byte, data []byte) SecurityBuffer {

	buffer := SecurityBuffer{}
	// Set the security buffer
	buffer.BufferLength = uint16(len(data))
	buffer.BufferAllocatedLength = uint16(len(data))
	buffer.StartOffset = uint32(len(*ntlmRawData))
	buffer.RawData = make([]byte, buffer.BufferAllocatedLength)

	*ntlmRawData = append(*ntlmRawData, data...)

	return buffer
}

func (sbuf *SecurityBuffer) ToBytes() []byte {
	buffer := []byte{}

	bufferLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bufferLengthBytes, sbuf.BufferLength)
	buffer = append(buffer, bufferLengthBytes...)

	bufferAllocatedLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bufferAllocatedLengthBytes, sbuf.BufferAllocatedLength)
	buffer = append(buffer, bufferAllocatedLengthBytes...)

	bufferStartOffsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bufferStartOffsetBytes, sbuf.StartOffset)
	buffer = append(buffer, bufferStartOffsetBytes...)

	buffer = append(buffer, sbuf.RawData...)

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
	str.WriteString(fmt.Sprintf("\tData String                  : %s\n", windows.ByteSliceToString(sbuf.RawData[:])))

	return str.String()
}
