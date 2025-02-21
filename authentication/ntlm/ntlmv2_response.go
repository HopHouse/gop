package ntlm

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type NTLMv2ResponseAttribute struct {
	Type   uint16
	Length int
	Data   []byte
}

func ReadNTLMv2ResponseAttribute(data []byte, start int) NTLMv2ResponseAttribute {
	buf := NTLMv2ResponseAttribute{
		Type:   binary.LittleEndian.Uint16(data[start : start+2]),
		Length: int(binary.LittleEndian.Uint16(data[start+2 : start+4])),
		Data:   []byte{},
	}
	buf.Data = data[start+4 : start+4+buf.Length]

	return buf
}

type NTLMv2Response struct {
	NTProofStr        []byte
	ResponseVersion   byte
	HiResponseVersion byte
	Timestamp         []byte
	Challenge         []byte
	Restrictions      NTLMv2ResponseAttribute
	ChannelBindings   NTLMv2ResponseAttribute
	TargetName        NTLMv2ResponseAttribute
}

func (resp *NTLMv2Response) Read(data []byte) {
	resp.NTProofStr = data[0:16]
	resp.ResponseVersion = data[16]
	resp.HiResponseVersion = data[17]
	resp.Timestamp = data[23:32]
	resp.Challenge = data[32:40]
	resp.Restrictions = ReadNTLMv2ResponseAttribute(data, 44)
	resp.ChannelBindings = ReadNTLMv2ResponseAttribute(data, 44+4+resp.Restrictions.Length)
	resp.TargetName = ReadNTLMv2ResponseAttribute(data, 44+4+4+resp.Restrictions.Length+resp.ChannelBindings.Length)
}

func (resp *NTLMv2Response) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("NTProofStr      	: 0x%x\n", resp.NTProofStr))
	str.WriteString(fmt.Sprintf("ResponseVersion  	: %d\n", resp.ResponseVersion))
	str.WriteString(fmt.Sprintf("HiResponseVersion	: %d\n", resp.HiResponseVersion))
	str.WriteString(fmt.Sprintf("Timestamp     		: %s\n", resp.Timestamp))
	str.WriteString(fmt.Sprintf("Challenge     		: 0x%x\n", resp.Challenge))
	str.WriteString(fmt.Sprintf("Restrictions   	: 0x%x\n", resp.Restrictions.Data))
	str.WriteString(fmt.Sprintf("ChannelBindings 	: 0x%x\n", resp.ChannelBindings.Data))
	str.WriteString(fmt.Sprintf("TargetName		: %v\n", string(resp.TargetName.Data)))

	return str.String()
}
