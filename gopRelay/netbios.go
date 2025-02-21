package gopRelay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const NETBIOS_SESSION_MESSAGE = 0x00
const NETBIOS_SESSION_REQUEST = 0x81
const NETBIOS_POSITIVE_SESSION_RESPONSE = 0x82
const NETBIOS_NEGATIVE_SESSION_RESPONSE = 0x83
const NETBIOS_RETARGET_SESSION_RESPONSE = 0x84
const NETBIOS_SESSION_KEEP_ALIVE = 0x85

type NetBiosPacket struct {
	MessageType byte
	Length      []byte
}

func (n *NetBiosPacket) Read(data []byte) error {
	n.MessageType = data[0]
	switch n.MessageType {
	case NETBIOS_SESSION_MESSAGE:
	case NETBIOS_SESSION_REQUEST:
	case NETBIOS_POSITIVE_SESSION_RESPONSE:
	case NETBIOS_NEGATIVE_SESSION_RESPONSE:
	case NETBIOS_RETARGET_SESSION_RESPONSE:
	case NETBIOS_SESSION_KEEP_ALIVE:
		break
	default:
		err := fmt.Errorf("netBIOS : Could not parse the n.MessageType : %x", n.MessageType)
		return err
	}
	n.Length = data[1:4]

	return nil
}

func (n *NetBiosPacket) SetLength(length uint32) {
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)
	n.Length = lengthBytes[1:4]
}

func (n *NetBiosPacket) ToBytes() []byte {
	var data bytes.Buffer

	binary.Write(&data, binary.LittleEndian, n.MessageType)
	binary.Write(&data, binary.BigEndian, n.Length)

	return data.Bytes()
}

func (n *NetBiosPacket) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("MessageType                    : %x (%d)\n", n.MessageType, n.MessageType))
	str.WriteString(fmt.Sprintf("Length                         : %x (%d)\n", n.Length, n.Length))

	return str.String()
}
