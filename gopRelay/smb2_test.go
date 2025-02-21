package gopRelay

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSMB2HeaderSync(t *testing.T) {

	smb2HeaderReference := []byte{0xfe, 0x53, 0x4d, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

	smb2Header := &SMB2_HEADER_SYNC{
		ProtocolID:    []byte{0xFE, 'S', 'M', 'B'},
		StructureSize: 64,
		CreditCharge:  0,
		NT_STATUS:     0x00000000, // STATUS_SUCCESS
		Command:       SMB2_COM_NEGOTIATE,
		Credits:       1,
		Flags:         0x00000001, // This is a responnse
		NextCommand:   0x00000000,
		MessageID:     0x00000000,
		Reserved:      0x00000000,
		TreeID:        0x00000000,
		SessionID:     0x0000000000000000,
		Signature:     [16]byte{},
	}

	smb2HeaderRead := &SMB2_HEADER_SYNC{}
	err := smb2HeaderRead.Read(smb2HeaderReference)
	if err != nil {
		t.Fatal(err)
	}

	compareStatus, compareStr := CompareBytesSlices(smb2Header.ToBytes(), smb2HeaderRead.ToBytes())
	if compareStatus != 0 {
		errorStr := bytes.NewBuffer([]byte{})
		fmt.Fprint(errorStr, smb2HeaderRead.ToString())
		fmt.Fprint(errorStr, compareStr)
		t.Fatal(errorStr.String())
	}

	compareStatus, compareStr = CompareBytesSlices(smb2HeaderReference, smb2Header.ToBytes())
	if compareStatus != 0 {
		errorStr := bytes.NewBuffer([]byte{})
		fmt.Fprint(errorStr, smb2Header.ToString())
		fmt.Fprint(errorStr, compareStr)
		t.Fatal(errorStr.String())
	}

	compareStatus, compareStr = CompareBytesSlices(smb2HeaderReference, smb2HeaderRead.ToBytes())
	if compareStatus != 0 {
		errorStr := bytes.NewBuffer([]byte{})
		fmt.Fprint(errorStr, smb2Header.ToString())
		fmt.Fprint(errorStr, compareStr)
		t.Fatal(errorStr.String())
	}
}

func TestSMB2_Negotiate_Protocol_Response(t *testing.T) {
	//
	// Wireshark Packet Description
	//

	// NetBIOS Session Service
	//     Message Type: Session message (0x00)
	//     Length: 176
	// SMB2 (Server Message Block Protocol version 2)
	//     SMB2 Header
	//         ProtocolId: 0xfe534d42
	//         Header Length: 64
	//         Credit Charge: 0
	//         NT Status: STATUS_SUCCESS (0x00000000)
	//         Command: Negotiate Protocol (0)
	//         Credits granted: 1
	//         Flags: 0x00000001, Response
	//             .... .... .... .... .... .... .... ...1 = Response: This is a RESPONSE
	//             .... .... .... .... .... .... .... ..0. = Async command: This is a SYNC command
	//             .... .... .... .... .... .... .... .0.. = Chained: This pdu is NOT a chained command
	//             .... .... .... .... .... .... .... 0... = Signing: This pdu is NOT signed
	//             .... .... .... .... .... .... .000 .... = Priority: This pdu does NOT contain a PRIORITY
	//             ...0 .... .... .... .... .... .... .... = DFS operation: This is a normal operation
	//             ..0. .... .... .... .... .... .... .... = Replay operation: This is NOT a replay operation
	//         Chain Offset: 0x00000000
	//         Message ID: 0
	//         Reserved: 0x00000000
	//         Tree Id: 0x00000000
	//         Session Id: 0x0000000000000000
	//         Signature: 00000000000000000000000000000000
	//         [Response to: 233]
	//         [Time from request: 0.000935000 seconds]
	//     Negotiate Protocol Response (0x00)
	//         [Preauth Hash: cb6cea8107813140c65c4633d498e51b930745a13b7bcdc6c42a4bb19601db25a36c3df1dc7ad1c3d1eebdd7de388184f484ce0bf49834c32800e69c745a4f5b]
	//         StructureSize: 0x0041
	//             0000 0000 0100 000. = Fixed Part Length: 32
	//             .... .... .... ...1 = Dynamic Part: True
	//         Security mode: 0x01, Signing enabled
	//             .... ...1 = Signing enabled: True
	//             .... ..0. = Signing required: False
	//         Dialect: SMB 2.0.2 (0x0202)
	//         Reserved: 0
	//         Server Guid: 70794e4e-6548-5669-7350-63746c526455
	//         Capabilities: 0x00000000
	//             .... .... .... .... .... .... .... ...0 = DFS: This host does NOT support DFS
	//             .... .... .... .... .... .... .... ..0. = LEASING: This host does NOT support LEASING
	//             .... .... .... .... .... .... .... .0.. = LARGE MTU: This host does NOT support LARGE_MTU
	//             .... .... .... .... .... .... .... 0... = MULTI CHANNEL: This host does NOT support MULTI CHANNEL
	//             .... .... .... .... .... .... ...0 .... = PERSISTENT HANDLES: This host does NOT support PERSISTENT HANDLES
	//             .... .... .... .... .... .... ..0. .... = DIRECTORY LEASING: This host does NOT support DIRECTORY LEASING
	//             .... .... .... .... .... .... .0.. .... = ENCRYPTION: This host does NOT support ENCRYPTION
	//             .... .... .... .... .... .... 0... .... = NOTIFICATIONS: This host does NOT support receiving NOTIFICATIONS
	//         Max Transaction Size: 65536
	//         Max Read Size: 65536
	//         Max Write Size: 65536
	//         Current Time: Feb 21, 2025 15:39:23.000000000 Romance Standard Time
	//         Boot Time: Feb 21, 2025 15:39:23.000000000 Romance Standard Time
	//         Blob Offset: 0x00000080
	//         Blob Length: 42
	//         Security Blob: 602806062b0601050502a01e301ca01a3018060a2b06010401823702021e060a2b06010401823702020a
	//             GSS-API Generic Security Service Application Program Interface
	//                 OID: 1.3.6.1.5.5.2 (SPNEGO - Simple Protected Negotiation)
	//                 Simple Protected Negotiation
	//                     negTokenInit
	//                         mechTypes: 2 items
	//                             MechType: 1.3.6.1.4.1.311.2.2.30 (NEGOEX - SPNEGO Extended Negotiation Security Mechanism)
	//                             MechType: 1.3.6.1.4.1.311.2.2.10 (NTLMSSP - Microsoft NTLM Security Support Provider)
	//         Reserved2: 0x00000000

	respGood := []byte{0x0, 0x0, 0x0, 0xb0, 0xfe, 0x53, 0x4d, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x41, 0x0, 0x1, 0x0, 0x2, 0x2, 0x0, 0x0, 0x4e, 0x4e, 0x79, 0x70, 0x48, 0x65, 0x69, 0x56, 0x73, 0x50, 0x63, 0x74, 0x6c, 0x52, 0x64, 0x55, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x80, 0x5f, 0xfe, 0x65, 0x6e, 0x84, 0xdb, 0x1, 0x80, 0x5f, 0xfe, 0x65, 0x6e, 0x84, 0xdb, 0x1, 0x80, 0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x60, 0x28, 0x6, 0x6, 0x2b, 0x6, 0x1, 0x5, 0x5, 0x2, 0xa0, 0x1e, 0x30, 0x1c, 0xa0, 0x1a, 0x30, 0x18, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0x1e, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

	// SMB
	// SMB2 Header
	smbHeader := &SMB2_HEADER_SYNC{
		ProtocolID:    []byte{0xFE, 'S', 'M', 'B'},
		StructureSize: 64,
		CreditCharge:  0,
		NT_STATUS:     0x00000000, // STATUS_SUCCESS
		Command:       SMB2_COM_NEGOTIATE,
		Credits:       1,
		Flags:         0x00000001, // This is a responnse
		NextCommand:   0x00000000,
		MessageID:     0x00000000,
		Reserved:      0x00000000,
		TreeID:        0x00000000,
		SessionID:     0x0000000000000000,
		Signature:     [16]byte{0x00, 0x00, 0x00, 0x00},
	}

	// SMB2 Negotiate Protcol Response
	sessionComNegotiateResponse := NewSMB2_NEGOTIATE_RESPONSE()

	resp, err := CreatePacket(smbHeader.ToBytes(), sessionComNegotiateResponse.ToBytes())
	if err != nil {
		t.Fatal(err)
	}

	compareStatus, compareStr := CompareBytesSlices(respGood, resp)
	if compareStatus != 0 {
		errorStr := bytes.NewBuffer([]byte{})

		fmt.Fprint(errorStr, smbHeader.ToString())
		fmt.Fprint(errorStr, sessionComNegotiateResponse.ToString())

		fmt.Fprint(errorStr, compareStr)

		t.Fatal(errorStr.String())
	}
}

func Test_SMB2_COM_SESSION_SETUP_RESPONSE_NTLSSP_Challenge_Response(t *testing.T) {

	// Session Setup Response (0x01)
	// [Preauth Hash: 8f5cfb0a050be676e6aa11ea7177b597f8aefe140ef5c44cfbdc29a9757fe96f4c1543b4685456275155c714e5514e59f74f1c3516219ae95e05a8eacaa607a4]
	// StructureSize: 0x0009
	//     0000 0000 0000 100. = Fixed Part Length: 4
	//     .... .... .... ...1 = Dynamic Part: True
	// Session Flags: 0x0000
	//     .... .... .... ...0 = Guest: False
	//     .... .... .... ..0. = Null: False
	//     .... .... .... .0.. = Encrypt: False
	// Blob Offset: 0x00000048
	// Blob Length: 271
	// Security Blob […]: a182010b30820107a0030a0101a10c060a2b06010401823702020aa281f10481ee4e544c4d53535000020000001e001e003800000015828ae22f9dafe6d995fa06000000000000000098009800560000000a00f4650000000f570049004e002d00460032005400440036004c00
	//     GSS-API Generic Security Service Application Program Interface
	//         Simple Protected Negotiation
	//             negTokenTarg
	//                 negResult: accept-incomplete (1)
	//                 supportedMech: 1.3.6.1.4.1.311.2.2.10 (NTLMSSP - Microsoft NTLM Security Support Provider)
	//                 responseToken […]: 4e544c4d53535000020000001e001e003800000015828ae22f9dafe6d995fa06000000000000000098009800560000000a00f4650000000f570049004e002d00460032005400440036004c005500540030005000520002001e00570049004e002d00460032005400440036004c
	//                 NTLM Secure Service Provider
	//                     NTLMSSP identifier: NTLMSSP
	//                     NTLM Message Type: NTLMSSP_CHALLENGE (0x00000002)
	//                     Target Name: WIN-F2TD6LUT0PR
	//                         Length: 30
	//                         Maxlen: 30
	//                         Offset: 56
	//                     Negotiate Flags: 0xe28a8215, Negotiate 56, Negotiate Key Exchange, Negotiate 128, Negotiate Version, Negotiate Target Info, Negotiate Extended Session Security, Target Type Server, Negotiate Always Sign, Negotiate NTLM key, Negotiate Sign
	//                         1... .... .... .... .... .... .... .... = Negotiate 56: Set
	//                         .1.. .... .... .... .... .... .... .... = Negotiate Key Exchange: Set
	//                         ..1. .... .... .... .... .... .... .... = Negotiate 128: Set
	//                         ...0 .... .... .... .... .... .... .... = Negotiate 0x10000000: Not set
	//                         .... 0... .... .... .... .... .... .... = Negotiate 0x08000000: Not set
	//                         .... .0.. .... .... .... .... .... .... = Negotiate 0x04000000: Not set
	//                         .... ..1. .... .... .... .... .... .... = Negotiate Version: Set
	//                         .... ...0 .... .... .... .... .... .... = Negotiate 0x01000000: Not set
	//                         .... .... 1... .... .... .... .... .... = Negotiate Target Info: Set
	//                         .... .... .0.. .... .... .... .... .... = Request Non-NT Session Key: Not set
	//                         .... .... ..0. .... .... .... .... .... = Negotiate 0x00200000: Not set
	//                         .... .... ...0 .... .... .... .... .... = Negotiate Identify: Not set
	//                         .... .... .... 1... .... .... .... .... = Negotiate Extended Session Security: Set
	//                         .... .... .... .0.. .... .... .... .... = Negotiate 0x00040000: Not set
	//                         .... .... .... ..1. .... .... .... .... = Target Type Server: Set
	//                         .... .... .... ...0 .... .... .... .... = Target Type Domain: Not set
	//                         .... .... .... .... 1... .... .... .... = Negotiate Always Sign: Set
	//                         .... .... .... .... .0.. .... .... .... = Negotiate 0x00004000: Not set
	//                         .... .... .... .... ..0. .... .... .... = Negotiate OEM Workstation Supplied: Not set
	//                         .... .... .... .... ...0 .... .... .... = Negotiate OEM Domain Supplied: Not set
	//                         .... .... .... .... .... 0... .... .... = Negotiate Anonymous: Not set
	//                         .... .... .... .... .... .0.. .... .... = Negotiate 0x00000400: Not set
	//                         .... .... .... .... .... ..1. .... .... = Negotiate NTLM key: Set
	//                         .... .... .... .... .... ...0 .... .... = Negotiate 0x00000100: Not set
	//                         .... .... .... .... .... .... 0... .... = Negotiate Lan Manager Key: Not set
	//                         .... .... .... .... .... .... .0.. .... = Negotiate Datagram: Not set
	//                         .... .... .... .... .... .... ..0. .... = Negotiate Seal: Not set
	//                         .... .... .... .... .... .... ...1 .... = Negotiate Sign: Set
	//                         .... .... .... .... .... .... .... 0... = Request 0x00000008: Not set
	//                         .... .... .... .... .... .... .... .1.. = Request Target: Set
	//                         .... .... .... .... .... .... .... ..0. = Negotiate OEM: Not set
	//                         .... .... .... .... .... .... .... ...1 = Negotiate UNICODE: Set
	//                     NTLM Server Challenge: 2f9dafe6d995fa06
	//                     Reserved: 0000000000000000
	//                     Target Info
	//                         Length: 152
	//                         Maxlen: 152
	//                         Offset: 86
	//                         Attribute: NetBIOS domain name: WIN-F2TD6LUT0PR
	//                             Target Info Item Type: NetBIOS domain name (0x0002)
	//                             Target Info Item Length: 30
	//                             NetBIOS Domain Name: WIN-F2TD6LUT0PR
	//                         Attribute: NetBIOS computer name: WIN-F2TD6LUT0PR
	//                             Target Info Item Type: NetBIOS computer name (0x0001)
	//                             Target Info Item Length: 30
	//                             NetBIOS Computer Name: WIN-F2TD6LUT0PR
	//                         Attribute: DNS domain name: WIN-F2TD6LUT0PR
	//                             Target Info Item Type: DNS domain name (0x0004)
	//                             Target Info Item Length: 30
	//                             DNS Domain Name: WIN-F2TD6LUT0PR
	//                         Attribute: DNS computer name: WIN-F2TD6LUT0PR
	//                             Target Info Item Type: DNS computer name (0x0003)
	//                             Target Info Item Length: 30
	//                             DNS Computer Name: WIN-F2TD6LUT0PR
	//                         Attribute: Timestamp
	//                             Target Info Item Type: Timestamp (0x0007)
	//                             Target Info Item Length: 8
	//                             Timestamp: Feb 19, 2025 23:17:05.129154500 Romance Standard Time
	//                         Attribute: End of list
	//                             Target Info Item Type: End of list (0x0000)
	//                             Target Info Item Length: 0
	//                     Version 10.0 (Build 26100); NTLM Current Revision 15
	//                         Major Version: 10
	//                         Minor Version: 0
	//                         Build Number: 26100
	//                         NTLM Current Revision: 15

	packetReference := []byte{0x9, 0x0, 0x0, 0x0, 0x48, 0x0, 0xf, 0x1, 0xa1, 0x82, 0x1, 0xb, 0x30, 0x82, 0x1, 0x7, 0xa0, 0x3, 0xa, 0x1, 0x1, 0xa1, 0xc, 0x6, 0xa, 0x2b, 0x6, 0x1, 0x4, 0x1, 0x82, 0x37, 0x2, 0x2, 0xa, 0xa2, 0x81, 0xf1, 0x4, 0x81, 0xee, 0x4e, 0x54, 0x4c, 0x4d, 0x53, 0x53, 0x50, 0x0, 0x2, 0x0, 0x0, 0x0, 0x1e, 0x0, 0x1e, 0x0, 0x38, 0x0, 0x0, 0x0, 0x15, 0x82, 0x8a, 0xe2, 0x2f, 0x9d, 0xaf, 0xe6, 0xd9, 0x95, 0xfa, 0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x98, 0x0, 0x98, 0x0, 0x56, 0x0, 0x0, 0x0, 0xa, 0x0, 0xf4, 0x65, 0x0, 0x0, 0x0, 0xf, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x2, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x1, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x4, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x3, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x7, 0x0, 0x8, 0x0, 0x99, 0x3b, 0xdf, 0x1, 0x1c, 0x83, 0xdb, 0x1, 0x0, 0x0, 0x0, 0x0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	packet := SMB2_COM_SESSION_SETUP_RESPONSE{}
	packet.Read(packetReference)

	compareStatus, compareStr := CompareBytesSlices(packetReference, packet.ToBytes())
	if compareStatus != 0 {
		t.Fatal(compareStr)
	}
}
