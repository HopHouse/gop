package gopRelay

import (
	"net"
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

type SMB_COM_NEGOTIATE_REQUEST struct {
	Parameters     byte
	WordCount      byte
	Data_ByteCount byte
}

func (s *SMB_COM_NEGOTIATE_REQUEST) Read(data []byte) {

}

func (s *SMB_COM_NEGOTIATE_REQUEST) ToString() string {
	return "Not implemented"
}

func (s *SMB_COM_NEGOTIATE_REQUEST) ToBytes() []byte {
	return []byte{}
}

type SMB_COM_NEGOTIATE_RESPONSE struct {
}

func (s *SMB_COM_NEGOTIATE_RESPONSE) Read(data []byte) {

}

func (s *SMB_COM_NEGOTIATE_RESPONSE) ToString() string {
	return "Not implemented"
}

func (s *SMB_COM_NEGOTIATE_RESPONSE) ToBytes() []byte {
	return []byte{}
}

func SMB_COM_NEGOTIATE_Receive(conn *net.TCPConn) error {

	return nil
}
