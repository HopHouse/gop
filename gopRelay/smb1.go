package gopRelay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/c083583f-1a8f-4afe-a742-6ee08ffeb8cf

// // Description:
// This diagram illustrates the sequence of messages exchanged between a client and a server in the Server Message Block (SMB) protocol during the session setup phase.
//
// 1. **SMB_COM_NEGOTIATE Request (Step 1)**: The client initiates the process by sending an SMB_COM_NEGOTIATE request to the server, indicating its intent to establish a connection and negotiate capabilities.
//
// 2. **SMB_COM_NEGOTIATE Response (Step 2)**: In response, the server sends back an SMB_COM_NEGOTIATE response, which outlines the supported SMB protocol versions and options.
//
// 3. **NTLM Negotiate Message (Step 3)**: The client proceeds with the SMB_COM_SESSION_SETUP_ANDX Request 1, including an NTLM Negotiate Message. This request signals the client's authentication method preferences.
//
// 4. **NTLM Challenge Message (Step 4)**: The server replies with SMB_COM_SESSION_SETUP_ANDX Response 1, containing an NTLM Challenge Message, which is necessary for the authentication process.
//
// 5. **NTLM Authenticate Message (Step 5)**: The client responds with SMB_COM_SESSION_SETUP_ANDX Request 2, which includes an NTLM Authenticate Message to confirm its identity.
//
// 6. **NTLM Authenticate Response (Step 6)**: Finally, the server sends SMB_COM_SESSION_SETUP_ANDX Response 2, completing the authentication process and establishing a session.
//
// This sequence is crucial for ensuring secure communication between the client and server, utilizing the NTLM authentication mechanism within the SMB protocol.

type Packet interface {
	Read([]byte) error
	ToString() string
	ToBytes() []byte
}

type SMB1_REQUEST_HEADER struct {
	Header        []byte // 0xFF, 'S', 'M', 'B'
	Command       uint8
	NT_STATUS     uint32
	Flags         uint8
	Flags2        uint16
	ProcessIDHigh uint8
	Signature     uint64
	_             uint16
	ProcessID     uint16
	TreeID        uint16
	UserID        uint16
	MultiplexID   uint16
}

func (s *SMB1_REQUEST_HEADER) Read(data []byte) error {
	s.Header = data[0:4]

	s.Command = data[4]
	switch s.Command {
	case SMB_COM_CREATE_DIRECTORY:
	case SMB_COM_DELETE_DIRECTORY:
	case SMB_COM_CLOSE:
	case SMB_COM_DELETE:
	case SMB_COM_RENAME:
	case SMB_COM_TRANSACTION:
	case SMB_COM_ECHO:
	case SMB_COM_OPEN_ANDX:
	case SMB_COM_READ_ANDX:
	case SMB_COM_WRITE_ANDX:
	case SMB_COM_TRANSACTION2:
	case SMB_COM_NEGOTIATE:
	case SMB_COM_SESSION_SETUP_ANDX:
	case SMB_COM_TREE_CONNECT_ANDX:
	case SMB_COM_NT_TRANSACT:
	case SMB_COM_NT_CREATE_ANDX:
		break
	default:
		err := fmt.Errorf("SMB1 : Could not parse the s.Command : %x", s.Command)
		return err
	}

	s.NT_STATUS = binary.LittleEndian.Uint32(data[5:9])

	s.Flags = data[9]
	s.Flags2 = binary.LittleEndian.Uint16(data[10:12])

	s.ProcessIDHigh = data[13]
	s.Signature = binary.LittleEndian.Uint64(data[14:22])

	_ = binary.LittleEndian.Uint16(data[22:26])
	s.TreeID = binary.LittleEndian.Uint16(data[26:28])
	s.ProcessID = binary.LittleEndian.Uint16(data[28:30])
	s.UserID = binary.LittleEndian.Uint16(data[30:32])
	s.MultiplexID = binary.LittleEndian.Uint16(data[32:34])

	return nil
}

func (s *SMB1_REQUEST_HEADER) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("Header                         : 0x%x%s\n", s.Header[0], string(s.Header[1:])))
	str.WriteString(fmt.Sprintf("Command                        : 0x%x (%s)\n", s.Command, SMB_COMMAND_NAMES[s.Command]))
	str.WriteString(fmt.Sprintf("NT_STATUS                      : 0x%x (%d)\n", s.NT_STATUS, s.NT_STATUS))
	str.WriteString(fmt.Sprintf("Flags                          : 0x%x\n", s.Flags))
	str.WriteString(fmt.Sprintf("Flags2                         : 0x%x\n", s.Flags2))
	str.WriteString(fmt.Sprintf("ProcessID High                 : 0x%x\n", s.ProcessIDHigh))
	str.WriteString(fmt.Sprintf("Pad or Security Signature      : 0x%x\n", s.Signature))
	str.WriteString(fmt.Sprintf("TreeID                         : 0x%x (%d)\n", s.TreeID, s.TreeID))
	str.WriteString(fmt.Sprintf("ProcessID                      : 0x%x (%d)\n", s.ProcessID, s.ProcessID))
	str.WriteString(fmt.Sprintf("UserID                         : 0x%x (%d)\n", s.UserID, s.UserID))
	str.WriteString(fmt.Sprintf("MultiplexID                    : 0x%x (%d)\n", s.MultiplexID, s.MultiplexID))

	return str.String()
}

func (s *SMB1_REQUEST_HEADER) ToBytes() []byte {
	return []byte{}
}

type SMB1_NEGOTIATE_REQUEST struct {
	WordCount uint8
	ByteCount uint16
	Dialects  []struct {
		BufferFormat uint8
		Name         []byte
	}
}

func (s *SMB1_NEGOTIATE_REQUEST) Read(data []byte) error {
	s.WordCount = data[0]
	s.ByteCount = binary.LittleEndian.Uint16(data[1:3])

	dialectsSlice := bytes.Split(data[3:], []byte{0x00})
	for _, dialect := range dialectsSlice {
		if len(dialect) == 0 {
			break
		}
		s.Dialects = append(s.Dialects, struct {
			BufferFormat uint8
			Name         []byte
		}{
			BufferFormat: dialect[0],
			Name:         dialect[1:],
		})
	}

	return nil
}

func (s *SMB1_NEGOTIATE_REQUEST) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("WordCount                      : 0x%x (%d)\n", s.WordCount, s.WordCount))
	str.WriteString(fmt.Sprintf("ByteCount                      : 0x%x (%d)\n", s.ByteCount, s.ByteCount))

	str.WriteString(fmt.Sprintf("Dialect (%d)                    :\n", len(s.Dialects)))
	for _, dialect := range s.Dialects {
		str.WriteString(fmt.Sprintf("\t- Format 0x%x : %s\n", dialect.BufferFormat, dialect.Name))
	}

	return str.String()
}

func (s *SMB1_NEGOTIATE_REQUEST) ToBytes() []byte {
	logger.Fatalln("SMB2_HEADER_SYNC.ToBytes() : Not implemented")
	return []byte{}
}
