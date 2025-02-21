package gopRelay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
	"time"

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

type SMB2_HEADER_SYNC struct {
	ProtocolID    []byte // 0xFE, 'S', 'M', 'B'
	StructureSize uint16
	CreditCharge  uint16
	NT_STATUS     uint32
	Command       uint16
	Credits       uint16
	Flags         uint32
	NextCommand   uint32
	MessageID     uint64
	Reserved      uint32
	TreeID        uint32
	SessionID     uint64
	Signature     [16]byte
}

func (s *SMB2_HEADER_SYNC) Read(data []byte) error {

	s.ProtocolID = data[0:4]

	s.StructureSize = binary.LittleEndian.Uint16(data[4:6])
	if s.StructureSize != 64 {
		err := fmt.Errorf("SMB2 : Header length must be greater than 64 bytes. Actual size is %d", s.StructureSize)
		return err
	}

	s.CreditCharge = binary.LittleEndian.Uint16(data[6:8])
	s.NT_STATUS = binary.LittleEndian.Uint32(data[8:12])
	s.Command = binary.LittleEndian.Uint16(data[12:14])
	switch s.Command {
	case SMB2_COM_NEGOTIATE:
	case SMB2_COM_SESSION_SETUP:
	case SMB2_COM_LOGOFF:
	case SMB2_COM_TREE_CONNECT:
	case SMB2_COM_TREE_DISCONNECT:
	case SMB2_COM_CREATE:
	case SMB2_COM_CLOSE:
	case SMB2_COM_FLUSH:
	case SMB2_COM_READ:
	case SMB2_COM_WRITE:
	case SMB2_COM_LOCK:
	case SMB2_COM_IOCTL:
	case SMB2_COM_CANCEL:
	case SMB2_COM_ECHO:
	case SMB2_COM_QUERY_DIRECTORY:
	case SMB2_COM_CHANGE_NOTIFY:
	case SMB2_COM_QUERY_INFO:
	case SMB2_COM_SET_INFO:
	case SMB2_COM_OPLOCK_BREAK:
		break
	default:
		err := fmt.Errorf("SMB2_HEADER_SYNC : Could not parse the s.NT_STATUS : %x", s.Command)
		return err

	}

	s.Credits = binary.LittleEndian.Uint16(data[14:16])
	s.Flags = binary.LittleEndian.Uint32(data[16:20])
	s.NextCommand = binary.LittleEndian.Uint32(data[20:24])
	s.MessageID = binary.LittleEndian.Uint64(data[24:32])
	s.Reserved = binary.LittleEndian.Uint32(data[32:36])
	s.TreeID = binary.LittleEndian.Uint32(data[36:40])
	s.SessionID = binary.LittleEndian.Uint64(data[40:48])
	s.Signature = [16]byte(data[48:64])

	return nil
}

func (s *SMB2_HEADER_SYNC) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("Header                         : 0x%x%s\n", s.ProtocolID[0], string(s.ProtocolID[1:])))
	str.WriteString(fmt.Sprintf("StructureSize                  : 0x%x (%d)\n", s.StructureSize, s.StructureSize))
	str.WriteString(fmt.Sprintf("CreditCharge                   : 0x%x (%d)\n", s.CreditCharge, s.CreditCharge))
	str.WriteString(fmt.Sprintf("NT_STATUS                      : 0x%x (%d)\n", s.NT_STATUS, s.NT_STATUS))
	str.WriteString(fmt.Sprintf("Command                        : 0x%x (%s)\n", s.Command, SMB2_COMMAND_NAMES[s.Command]))
	str.WriteString(fmt.Sprintf("Credits                        : 0x%x (%d)\n", s.Credits, s.Credits))
	str.WriteString(fmt.Sprintf("Flags                          : 0x%x (%d)\n", s.Flags, s.Flags))
	str.WriteString(fmt.Sprintf("Next Command                   : 0x%x (%d)\n", s.NextCommand, s.NextCommand))
	str.WriteString(fmt.Sprintf("MessageID                      : 0x%x (%d)\n", s.MessageID, s.MessageID))
	str.WriteString(fmt.Sprintf("Reserved                       : 0x%x (%d)\n", s.Reserved, s.Reserved))
	str.WriteString(fmt.Sprintf("TreeID                         : 0x%x (%d)\n", s.TreeID, s.TreeID))
	str.WriteString(fmt.Sprintf("SessionID                      : 0x%x (%d)\n", s.SessionID, s.SessionID))
	str.WriteString(fmt.Sprintf("Signature                      : 0x%x\n", s.Signature))

	return str.String()
}

func (s *SMB2_HEADER_SYNC) ToBytes() []byte {
	var data bytes.Buffer

	binary.Write(&data, binary.LittleEndian, []byte{0xFE, 'S', 'M', 'B'})
	binary.Write(&data, binary.LittleEndian, s.StructureSize)
	binary.Write(&data, binary.LittleEndian, s.CreditCharge)
	binary.Write(&data, binary.LittleEndian, s.NT_STATUS)
	binary.Write(&data, binary.LittleEndian, s.Command)
	binary.Write(&data, binary.LittleEndian, s.Credits)
	binary.Write(&data, binary.LittleEndian, s.Flags)
	binary.Write(&data, binary.LittleEndian, s.NextCommand)
	binary.Write(&data, binary.LittleEndian, s.MessageID)
	binary.Write(&data, binary.LittleEndian, s.Reserved)
	binary.Write(&data, binary.LittleEndian, s.TreeID)
	binary.Write(&data, binary.LittleEndian, s.SessionID)
	binary.Write(&data, binary.LittleEndian, s.Signature)

	return data.Bytes()
}

func (s *SMB2_HEADER_SYNC) ToBytesOld() []byte {
	data := []byte{}

	// Header
	data = append(data, []byte{0xFE, 'S', 'M', 'B'}...)

	StructureSizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(StructureSizeBytes, s.StructureSize)
	data = append(data, StructureSizeBytes...)

	CreditChargeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(CreditChargeBytes, s.CreditCharge)
	data = append(data, CreditChargeBytes...)

	NT_STATUSBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(NT_STATUSBytes, s.NT_STATUS)
	data = append(data, NT_STATUSBytes...)

	CommandBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(CommandBytes, s.Command)
	data = append(data, CommandBytes...)

	CreditsBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(CreditsBytes, s.Credits)
	data = append(data, CreditsBytes...)

	FlagsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(FlagsBytes, s.Flags)
	data = append(data, FlagsBytes...)

	NextCommandBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(NextCommandBytes, s.NextCommand)
	data = append(data, NextCommandBytes...)

	MessageIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(MessageIDBytes, s.MessageID)
	data = append(data, MessageIDBytes...)

	ReservedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(ReservedBytes, s.Reserved)
	data = append(data, ReservedBytes...)

	TreeIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(TreeIDBytes, s.TreeID)
	data = append(data, TreeIDBytes...)

	SessionIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(SessionIDBytes, s.SessionID)
	data = append(data, SessionIDBytes...)

	data = append(data, s.Signature[:]...)

	return data
}

type SMB2_NEGOTIATE_REQUEST struct {
	StructureSize            uint16
	DialectCount             uint16
	SecurityMode             uint16
	Reserved                 uint16
	Capabilities             uint32
	ClientGuid               []byte // 16 bytes
	NegotiateContextOffset   uint32
	NegotiateContextCount    uint16
	NegotiateContextReserved uint16
	Dialects                 []uint16
	Padding                  []byte
	NegotiateContextList     []SMB2_NEGOTIATE_CONTEXT
}

func (s *SMB2_NEGOTIATE_REQUEST) Read(data []byte) error {
	s.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	s.DialectCount = binary.LittleEndian.Uint16(data[2:4])
	s.SecurityMode = binary.LittleEndian.Uint16(data[4:6])
	s.Reserved = binary.LittleEndian.Uint16(data[6:8])
	s.Capabilities = binary.LittleEndian.Uint32(data[8:12])
	s.ClientGuid = data[12:28]
	s.NegotiateContextOffset = binary.LittleEndian.Uint32(data[28:32])
	s.NegotiateContextCount = binary.LittleEndian.Uint16(data[32:34])
	s.NegotiateContextReserved = binary.LittleEndian.Uint16(data[34:36])

	offset := 36
	for i := uint16(0); i < s.DialectCount; i++ {
		s.Dialects = append(s.Dialects, binary.LittleEndian.Uint16(data[offset:offset+2]))
		offset += 2
	}

	for i := offset % 8; i > 0; i-- {
		s.Padding = append(s.Padding, 0x00)
	}

	if slices.Contains(s.Dialects, 0x311) {
		negotiateOffset := s.NegotiateContextOffset

		for i := uint16(0); i < s.NegotiateContextCount; i++ {
			context := SMB2_NEGOTIATE_CONTEXT{}
			// Give all data and compute size after
			context.Read(data[negotiateOffset:])
			s.NegotiateContextList = append(s.NegotiateContextList, context)

			// Get size taken by the Context
			negotiateOffset += uint32(context.DataLength)

			// Add padding
			negotiateOffset += negotiateOffset % 8
		}

	} else {
		s.NegotiateContextList = []SMB2_NEGOTIATE_CONTEXT{}
	}

	return nil
}

func (s *SMB2_NEGOTIATE_REQUEST) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("StructureSize                  : %d\n", s.StructureSize))
	str.WriteString(fmt.Sprintf("DialectCount                   : 0x%x (%d)\n", s.DialectCount, s.DialectCount))

	// Security mode
	securityModeString := ""
	switch s.SecurityMode {
	case SMB2_NEGOTIATE_SIGNING_ENABLED:
		securityModeString = "SIGNING ENABLED"
	case SMB2_NEGOTIATE_SIGNING_REQUIRED:
		securityModeString = "SIGNING REQUIRED"
	default:
		securityModeString = "Unknown"
	}
	str.WriteString(fmt.Sprintf("SecurityMode                   : 0x%x (%s)\n", s.SecurityMode, securityModeString))
	str.WriteString(fmt.Sprintf("Reserved                       : 0x%x (%d)\n", s.Reserved, s.Reserved))

	str.WriteString(fmt.Sprintf("Capabilities                   : 0x%x (%d)\n", s.Capabilities, s.Capabilities))
	str.WriteString(fmt.Sprintf("ClientGuid                     : 0x%x (%s)\n", s.ClientGuid, s.ClientGuid))

	str.WriteString(fmt.Sprintf("NegotiateContextOffset         : 0x%x (%d)\n", s.NegotiateContextOffset, s.NegotiateContextOffset))
	str.WriteString(fmt.Sprintf("NegotiateContextCount          : 0x%x (%d)\n", s.NegotiateContextCount, s.NegotiateContextCount))
	str.WriteString(fmt.Sprintf("NegotiateContextReserved       : 0x%x (%d)\n", s.NegotiateContextReserved, s.NegotiateContextReserved))

	str.WriteString(fmt.Sprintf("Dialect (%d)                    :\n", len(s.Dialects)))
	for _, dialect := range s.Dialects {
		str.WriteString(fmt.Sprintf("\t0x%x (%s)\n", dialect, SMB2_DIALECT_NAMES[dialect]))
	}

	str.WriteString(fmt.Sprintf("Reserved                       : 0x%x (%d)\n", s.Reserved, s.Reserved))
	str.WriteString(fmt.Sprintf("NegotiateContext (%d)           :\n", len(s.NegotiateContextList)))
	for _, negotiate := range s.NegotiateContextList {
		str.WriteString(negotiate.ToString())
	}

	return str.String()
}

func (s *SMB2_NEGOTIATE_REQUEST) ToBytes() []byte {
	logger.Fatalln("SMB2_HEADER_SYNC.ToBytes() : Not implemented")
	return []byte{}
}

type SMB2_NEGOTIATE_CONTEXT struct {
	ContextType uint16
	DataLength  uint16
	Reserved    uint16
	Data        []byte
}

func (s *SMB2_NEGOTIATE_CONTEXT) Read(data []byte) error {
	s.ContextType = binary.LittleEndian.Uint16(data[0:2])
	if _, exist := SMB2_CONTEXT_NAMES[s.ContextType]; !exist {
		err := fmt.Errorf("SMB2_NEGOTIATE_CONTEXT : Could not parse the s.ContextType : 0x%x", s.ContextType)
		return err
	}

	s.DataLength = binary.LittleEndian.Uint16(data[2:4])

	s.Reserved = binary.LittleEndian.Uint16(data[6:8])

	computedDataLength := uint16(len(data) - 8)
	if s.DataLength < computedDataLength {
		err := fmt.Errorf("SMB2_NEGOTIATE_CONTEXT : buffer data lower than the s.DataLentgh (%d) : %d", s.DataLength, computedDataLength)
		return err
	}

	s.Data = data[8 : 8+s.DataLength]

	return nil
}

func (s *SMB2_NEGOTIATE_CONTEXT) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("ContextType                    : 0x%x (%s)\n", s.ContextType, SMB2_CONTEXT_NAMES[s.ContextType]))
	str.WriteString(fmt.Sprintf("DataLength                     : 0x%x (%d)\n", s.DataLength, s.DataLength))
	str.WriteString(fmt.Sprintf("Reserved                       : 0x%x (%d)\n", s.Reserved, s.Reserved))
	str.WriteString(fmt.Sprintf("Data                           : 0x%x\n", s.Data))

	return str.String()
}

func (s *SMB2_NEGOTIATE_CONTEXT) ToBytes() []byte {
	logger.Fatalln("SMB2_NEGOTIATE_CONTEXT.ToBytes() : Not implemented")
	return []byte{}
}

type SMB2_NEGOTIATE_RESPONSE struct {
	StructureSize   uint16
	SecurityMode    uint16
	DialectRevision uint16
	Reserved        uint16
	ServerGuid      []byte // 16 bytes
	Capabilities    uint32

	MaxTransactSize uint32
	MaxReadSize     uint32
	MaxWriteSize    uint32
	SystemTime      uint64
	ServerStartTime uint64

	SecurityBufferOffset uint16
	SecurityBufferLength uint16

	NegotiateContextOffset uint32

	Buffer               []byte
	Padding              []byte
	NegotiateContextList []SMB2_NEGOTIATE_CONTEXT
}

func (s *SMB2_NEGOTIATE_RESPONSE) Read(data []byte) error {
	s.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	s.SecurityMode = binary.LittleEndian.Uint16(data[2:4])
	s.DialectRevision = binary.LittleEndian.Uint16(data[4:6])
	s.Reserved = binary.LittleEndian.Uint16(data[6:8])
	s.ServerGuid = data[8:24]
	s.Capabilities = binary.LittleEndian.Uint32(data[24:28])
	s.MaxTransactSize = binary.LittleEndian.Uint32(data[28:32])
	s.MaxReadSize = binary.LittleEndian.Uint32(data[32:36])
	s.MaxWriteSize = binary.LittleEndian.Uint32(data[36:40])
	s.SystemTime = binary.LittleEndian.Uint64(data[40:48])
	s.ServerStartTime = binary.LittleEndian.Uint64(data[48:56])
	s.SecurityBufferOffset = binary.LittleEndian.Uint16(data[56:58])
	s.SecurityBufferLength = binary.LittleEndian.Uint16(data[58:60])
	s.NegotiateContextOffset = binary.LittleEndian.Uint32(data[60:64])

	offset := s.SecurityBufferOffset
	s.Buffer = data[offset : offset+s.SecurityBufferLength]

	offset = offset + s.SecurityBufferLength
	for i := offset % 8; i > 0; i++ {
		s.Padding = append(s.Padding, 0x00)
	}

	if s.DialectRevision == 0x311 {
		negotiateOffset := s.NegotiateContextOffset

		for {
			if negotiateOffset > uint32(len(data[negotiateOffset:])) {
				break
			}

			context := SMB2_NEGOTIATE_CONTEXT{}
			// Give all data and compute size after
			context.Read(data[negotiateOffset:])
			s.NegotiateContextList = append(s.NegotiateContextList, context)

			// Get size taken by the Context
			negotiateOffset += uint32(context.DataLength)

			// Add padding
			negotiateOffset += negotiateOffset % 8
		}

	} else {
		s.NegotiateContextList = []SMB2_NEGOTIATE_CONTEXT{}
	}

	return nil
}

func (s *SMB2_NEGOTIATE_RESPONSE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("StructureSize                  : %d\n", s.StructureSize))
	// Security mode
	securityModeString := ""
	switch s.SecurityMode {
	case SMB2_NEGOTIATE_SIGNING_ENABLED:
		securityModeString = "SIGNING ENABLED"
	case SMB2_NEGOTIATE_SIGNING_REQUIRED:
		securityModeString = "SIGNING REQUIRED"
	default:
		securityModeString = "Unknown"
	}
	str.WriteString(fmt.Sprintf("SecurityMode                   : 0x%x (%s)\n", s.SecurityMode, securityModeString))
	str.WriteString(fmt.Sprintf("DialectRevision                : 0x%x (%d)\n", s.DialectRevision, s.DialectRevision))
	str.WriteString(fmt.Sprintf("Reserved                       : 0x%x (%d)\n", s.Reserved, s.Reserved))
	str.WriteString(fmt.Sprintf("ServerGuid                     : 0x%x (%s)\n", s.ServerGuid, s.ServerGuid))
	str.WriteString(fmt.Sprintf("Capabilities                   : 0x%x (%d)\n", s.Capabilities, s.Capabilities))

	str.WriteString(fmt.Sprintf("MaxTransctSize                 : 0x%x (%d)\n", s.MaxTransactSize, s.MaxTransactSize))
	str.WriteString(fmt.Sprintf("MaxReadSize                    : 0x%x (%d)\n", s.MaxReadSize, s.MaxReadSize))
	str.WriteString(fmt.Sprintf("MaxWriteSize                   : 0x%x (%d)\n", s.MaxWriteSize, s.MaxWriteSize))

	systemTime := time.Unix(int64(s.SystemTime), 0)
	str.WriteString(fmt.Sprintf("SystemTime                     : %s\n", systemTime.String()))

	serverStartTime := time.Unix(int64(s.ServerStartTime), 0)
	str.WriteString(fmt.Sprintf("ServerStartTime                : %s\n", serverStartTime.String()))

	str.WriteString(fmt.Sprintf("SecurityBufferOffset           : 0x%x (%d)\n", s.SecurityBufferOffset, s.SecurityBufferOffset))
	str.WriteString(fmt.Sprintf("SecurityBufferLength           : 0x%x (%d)\n", s.SecurityBufferLength, s.SecurityBufferLength))
	str.WriteString(fmt.Sprintf("NegotiationContextOffset       : 0x%x (%d)\n", s.NegotiateContextOffset, s.NegotiateContextOffset))

	str.WriteString(fmt.Sprintf("Buffer                         : 0x%x\n", s.Buffer))

	str.WriteString(fmt.Sprintf("Padding                        : 0x%x (%d)\n", s.Padding, s.Padding))
	str.WriteString(fmt.Sprintf("NegotiateContext (%d)          :\n", len(s.NegotiateContextList)))
	for _, negotiate := range s.NegotiateContextList {
		str.WriteString(negotiate.ToString())
	}

	return str.String()
}

func (s *SMB2_NEGOTIATE_RESPONSE) ToBytes() []byte {
	data := []byte{}

	StructureSizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(StructureSizeBytes, s.StructureSize)
	data = append(data, StructureSizeBytes...)

	SecurityModeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(SecurityModeBytes, s.SecurityMode)
	data = append(data, SecurityModeBytes...)

	DialectRevisionBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(DialectRevisionBytes, s.DialectRevision)
	data = append(data, DialectRevisionBytes...)

	ReservedBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(ReservedBytes, s.Reserved)
	data = append(data, ReservedBytes...)

	data = append(data, s.ServerGuid[0:max(16, len(s.ServerGuid))]...)

	CapabilitiesBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(CapabilitiesBytes, s.Capabilities)
	data = append(data, CapabilitiesBytes...)

	MaxTransactSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(MaxTransactSizeBytes, s.MaxTransactSize)
	data = append(data, MaxTransactSizeBytes...)

	MaxReadSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(MaxReadSizeBytes, s.MaxReadSize)
	data = append(data, MaxReadSizeBytes...)

	MaxWriteSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(MaxWriteSizeBytes, s.MaxWriteSize)
	data = append(data, MaxWriteSizeBytes...)

	SystemTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(SystemTimeBytes, s.SystemTime)
	data = append(data, SystemTimeBytes...)

	ServerStartTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(ServerStartTimeBytes, s.ServerStartTime)
	data = append(data, ServerStartTimeBytes...)

	SecurityBufferOffsetBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(SecurityBufferOffsetBytes, s.SecurityBufferOffset)
	data = append(data, SecurityBufferOffsetBytes...)

	SecurityBufferLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(SecurityBufferLengthBytes, s.SecurityBufferLength)
	data = append(data, SecurityBufferLengthBytes...)

	NegotiateContextOffsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(NegotiateContextOffsetBytes, s.NegotiateContextOffset)
	data = append(data, NegotiateContextOffsetBytes...)

	data = append(data, s.Buffer...)
	data = append(data, s.Padding...)

	for i := len(data); i < int(s.StructureSize); i++ {
		logger.Panicln("Adding padding")
		data = append(data, 0x00)
	}

	return data
}

func (s *SMB2_NEGOTIATE_RESPONSE) GetLength() uint16 {
	// TODO Verify if 64 by default and +1 if additioanl data

	// Structure Size
	length := uint16(64)
	length += s.SecurityBufferLength

	for _, context := range s.NegotiateContextList {
		length += context.DataLength
	}

	return length
}

func NewSMB2_NEGOTIATE_RESPONSE() *SMB2_NEGOTIATE_RESPONSE {
	resp := SMB2_NEGOTIATE_RESPONSE{}

	resp.StructureSize = 65
	resp.SecurityMode = SMB2_NEGOTIATE_SIGNING_ENABLED
	resp.DialectRevision = SMB2_DIALECT_202
	resp.Reserved = 0
	// TODO : Fix it
	// serverGuid, _ := ntlm.ByteSliceFromString("a81fbcd2-8dda-4361-80d2-4c852824f572")
	// resp.ServerGuid = serverGuid
	resp.ServerGuid = []byte{0xd2, 0xbc, 0x1f, 0xa8, 0xda, 0x8d, 0x61, 0x43, 0x80, 0xd2, 0x4c, 0x85, 0x28, 0x24, 0xf5, 0x72}
	resp.Capabilities = 0x00000007
	resp.MaxTransactSize = 8388608
	resp.MaxReadSize = 8388608
	resp.MaxWriteSize = 8388608
	// resp.SystemTime = uint64(time.Now().Unix())
	resp.SystemTime = binary.LittleEndian.Uint64([]byte{0x57, 0x40, 0x7e, 0xff, 0xce, 0x82, 0xdb, 0x1})
	resp.ServerStartTime = 0
	resp.SecurityBufferOffset = 0x00000080
	resp.SecurityBufferLength = 42

	// Buffer from wireshark :
	// Security Blob: 602806062b0601050502a01e301ca01a3018060a2b06010401823702021e060a2b06010401823702020a
	//    GSS-API Generic Security Service Application Program Interface
	//        OID: 1.3.6.1.5.5.2 (SPNEGO - Simple Protected Negotiation)
	//        Simple Protected Negotiation
	//            negTokenInit
	//                mechTypes: 2 items
	//                    MechType: 1.3.6.1.4.1.311.2.2.30 (NEGOEX - SPNEGO Extended Negotiation Security Mechanism)
	//                    MechType: 1.3.6.1.4.1.311.2.2.10 (NTLMSSP - Microsoft NTLM Security Support Provider)

	resp.Buffer = []byte{0x60, 0x28, 0x6, 0x6, 0x2b, 0x6, 0x1, 0x5, 0x5, 0x2, 0xa0, 0x1e, 0x30, 0x1c, 0xa0, 0x1a, 0x30, 0x18, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0x1e, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0xa}

	return &resp
}

type SMB2_COM_SESSION_SETUP_REQUEST struct {
	StructureSize        uint16
	Flags                byte
	SecurityMode         byte
	Capabilities         uint32
	Channel              uint32
	SecurityBufferOffset uint16
	SecurityBufferLength uint16
	PreviousSessionId    uint64
	Buffer               []byte
}

func (s *SMB2_COM_SESSION_SETUP_REQUEST) Read(data []byte) error {
	s.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	if s.StructureSize != 25 {
		err := fmt.Errorf("structureSize is not 25 : %d", s.StructureSize)
		return err
	}

	s.Flags = data[2]
	s.SecurityMode = data[3]
	s.Capabilities = binary.LittleEndian.Uint32(data[4:8])
	s.Channel = binary.LittleEndian.Uint32(data[8:12])
	s.SecurityBufferOffset = binary.LittleEndian.Uint16(data[12:14])
	s.SecurityBufferLength = binary.LittleEndian.Uint16(data[14:26])
	s.PreviousSessionId = binary.LittleEndian.Uint64(data[16:24])
	s.Buffer = data[24 : 24+s.SecurityBufferLength]

	return nil
}

func (s *SMB2_COM_SESSION_SETUP_REQUEST) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("StructureSize                  : %d\n", s.StructureSize))
	str.WriteString(fmt.Sprintf("Flags                          : 0x%x\n", s.Flags))
	// Security mode
	securityModeString := ""
	switch s.SecurityMode {
	case 0x01:
		securityModeString = "SIGNING ENABLED"
	case 0x02:
		securityModeString = "SIGNING REQUIRED"
	default:
		securityModeString = "Unknown"
	}
	str.WriteString(fmt.Sprintf("SecurityMode                   : 0x%x (%s)\n", s.SecurityMode, securityModeString))
	str.WriteString(fmt.Sprintf("Capabilities                   : 0x%x\n", s.Capabilities))
	str.WriteString(fmt.Sprintf("Channel                        : 0x%x\n", s.Channel))
	str.WriteString(fmt.Sprintf("SecurityBuffer Offset          : 0x%x\n", s.SecurityBufferOffset))
	str.WriteString(fmt.Sprintf("SecurityBuffer Length          : 0x%x (%d)\n", s.SecurityBufferLength, s.SecurityBufferLength))
	str.WriteString(fmt.Sprintf("Buffer                         : 0x%x\n", s.Buffer))

	return str.String()
}

func (s *SMB2_COM_SESSION_SETUP_REQUEST) ToBytes() []byte {
	logger.Fatalln("SMB2_COM_SESSION_SETUP_REQUEST.ToBytes() : Not implemented")
	return []byte{}
}

type SMB2_COM_SESSION_SETUP_RESPONSE struct {
	StructureSize        uint16
	SessionFlags         uint16
	SecurityBufferOffset uint16
	SecurityBufferLength uint16
	Buffer               []byte
}

func (s *SMB2_COM_SESSION_SETUP_RESPONSE) Read(data []byte) error {
	s.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	if s.StructureSize != 9 {
		err := fmt.Errorf("structureSize is not 9 : %d", s.StructureSize)
		return err
	}
	s.SessionFlags = binary.LittleEndian.Uint16(data[2:4])
	s.SecurityBufferOffset = binary.LittleEndian.Uint16(data[4:6])
	s.SecurityBufferLength = binary.LittleEndian.Uint16(data[6:8])
	s.Buffer = data[8:]

	return nil
}

func (s *SMB2_COM_SESSION_SETUP_RESPONSE) ToString() string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("StructureSize                  : %d\n", s.StructureSize))
	str.WriteString(fmt.Sprintf("SessionFlags                   : 0x%x\n", s.SessionFlags))
	str.WriteString(fmt.Sprintf("SecurityBuffer Offset          : 0x%x\n", s.SecurityBufferOffset))
	str.WriteString(fmt.Sprintf("SecurityBuffer Length          : 0x%x (%d)\n", s.SecurityBufferLength, s.SecurityBufferLength))
	str.WriteString(fmt.Sprintf("Buffer                         : 0x%x\n", s.Buffer))

	return str.String()
}

func (s *SMB2_COM_SESSION_SETUP_RESPONSE) ToBytes() []byte {
	var data bytes.Buffer

	binary.Write(&data, binary.LittleEndian, s.StructureSize)
	binary.Write(&data, binary.LittleEndian, s.SessionFlags)
	binary.Write(&data, binary.LittleEndian, s.SecurityBufferOffset)
	binary.Write(&data, binary.LittleEndian, s.SecurityBufferLength)
	binary.Write(&data, binary.LittleEndian, s.Buffer)

	return data.Bytes()
}

func (s *SMB2_COM_SESSION_SETUP_RESPONSE) GetLength() uint16 {
	return s.StructureSize + s.SecurityBufferLength
}

func NewSMB2_COM_SESSION_SETUP_RESPONSE(data []byte, offset uint16) *SMB2_COM_SESSION_SETUP_RESPONSE {
	resp := &SMB2_COM_SESSION_SETUP_RESPONSE{}

	resp.Buffer = []byte{0xa1, 0x82, 0x1, 0xb, 0x30, 0x82, 0x1, 0x7, 0xa0, 0x3, 0xa, 0x1, 0x1, 0xa1, 0xc, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0xa, 0xa2, 0x81, 0xf1, 0x4, 0x81, 0xee}
	resp.Buffer = append(resp.Buffer, data...)

	resp.StructureSize = 9
	resp.SessionFlags = 0
	resp.SecurityBufferOffset = 9 + offset - 1

	resp.SecurityBufferLength = uint16(len(resp.Buffer))

	return resp
}
