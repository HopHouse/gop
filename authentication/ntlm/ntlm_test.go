package ntlm

import (
	"testing"
)

func TestNTLMSSPNegitiateRequest(t *testing.T) {
	//
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

	packetReference := []byte{0x4e, 0x54, 0x4c, 0x4d, 0x53, 0x53, 0x50, 0x0, 0x1, 0x0, 0x0, 0x0, 0x97, 0x82, 0x8, 0xe2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0xf4, 0x65, 0x0, 0x0, 0x0, 0xf}

	NTLMNegotiate := NTLMSSP_NEGOTIATE{}
	err := NTLMNegotiate.Read(packetReference)
	if err != nil {
		t.Fatal(err)
	}

	equal, str := CompareBytesSlices(packetReference, NTLMNegotiate.ToBytes())
	if equal != 0 {
		t.Fatal(str)
	}
}

func TestNTLMSSPChallengeRequest(t *testing.T) {
	//
	// Security Blob […]: a182020d30820209a0030a0101a28201ec048201e84e544c4d535350000300000018001800820000003e013e019a00000002000200580000000e000e005a0000001a001a006800000010001000d8010000158288e20a00f4650000000fcb60bac4d825353456ed59130710c76c
	//     GSS-API Generic Security Service Application Program Interface
	//         Simple Protected Negotiation
	//             negTokenTarg
	//                 negResult: accept-incomplete (1)
	//                 responseToken […]: 4e544c4d535350000300000018001800820000003e013e019a00000002000200580000000e000e005a0000001a001a006800000010001000d8010000158288e20a00f4650000000fcb60bac4d825353456ed59130710c76c2e00610075006400690074006f007200570049004e
	//                 NTLM Secure Service Provider
	//                     NTLMSSP identifier: NTLMSSP
	//                     NTLM Message Type: NTLMSSP_AUTH (0x00000003)
	//                     Lan Manager Response: 000000000000000000000000000000000000000000000000
	//                         Length: 24
	//                         Maxlen: 24
	//                         Offset: 130
	//                     LMv2 Client Challenge: 0000000000000000
	//                     NTLM Response […]: 435f54967117ebdf70365302a420391d0101000000000000993bdf011c83db0162ca571c5eded6c60000000002001e00570049004e002d00460032005400440036004c005500540030005000520001001e00570049004e002d00460032005400440036004c0055005400300050
	//                         Length: 318
	//                         Maxlen: 318
	//                         Offset: 154
	//                         NTLMv2 Response […]: 435f54967117ebdf70365302a420391d0101000000000000993bdf011c83db0162ca571c5eded6c60000000002001e00570049004e002d00460032005400440036004c005500540030005000520001001e00570049004e002d00460032005400440036004c00550054003000
	//                             NTProofStr: 435f54967117ebdf70365302a420391d
	//                             Response Version: 1
	//                             Hi Response Version: 1
	//                             Z: 000000000000
	//                             Time: Feb 19, 2025 22:17:05.129154500 UTC
	//                             NTLMv2 Client Challenge: 62ca571c5eded6c6
	//                             Z: 00000000
	//                             Attribute: NetBIOS domain name: WIN-F2TD6LUT0PR
	//                                 NTLMV2 Response Item Type: NetBIOS domain name (0x0002)
	//                                 NTLMV2 Response Item Length: 30
	//                                 NetBIOS Domain Name: WIN-F2TD6LUT0PR
	//                             Attribute: NetBIOS computer name: WIN-F2TD6LUT0PR
	//                                 NTLMV2 Response Item Type: NetBIOS computer name (0x0001)
	//                                 NTLMV2 Response Item Length: 30
	//                                 NetBIOS Computer Name: WIN-F2TD6LUT0PR
	//                             Attribute: DNS domain name: WIN-F2TD6LUT0PR
	//                                 NTLMV2 Response Item Type: DNS domain name (0x0004)
	//                                 NTLMV2 Response Item Length: 30
	//                                 DNS Domain Name: WIN-F2TD6LUT0PR
	//                             Attribute: DNS computer name: WIN-F2TD6LUT0PR
	//                                 NTLMV2 Response Item Type: DNS computer name (0x0003)
	//                                 NTLMV2 Response Item Length: 30
	//                                 DNS Computer Name: WIN-F2TD6LUT0PR
	//                             Attribute: Timestamp
	//                                 NTLMV2 Response Item Type: Timestamp (0x0007)
	//                                 NTLMV2 Response Item Length: 8
	//                                 Timestamp: Feb 19, 2025 23:17:05.129154500 Romance Standard Time
	//                             Attribute: Flags
	//                                 NTLMV2 Response Item Type: Flags (0x0006)
	//                                 NTLMV2 Response Item Length: 4
	//                                 Flags: 0x00000002
	//                             Attribute: Restrictions
	//                                 NTLMV2 Response Item Type: Restrictions (0x0008)
	//                                 NTLMV2 Response Item Length: 48
	//                                 Restrictions: 30000000000000000000000000300000f530ea47e268c8d5b60082a2a7977fe0638e0e171a6397c51c722d3aee86ad46
	//                             Attribute: Channel Bindings
	//                                 NTLMV2 Response Item Type: Channel Bindings (0x000a)
	//                                 NTLMV2 Response Item Length: 16
	//                                 Channel Bindings: 00000000000000000000000000000000
	//                             Attribute: Target Name: cifs/192.168.71.7
	//                                 NTLMV2 Response Item Type: Target Name (0x0009)
	//                                 NTLMV2 Response Item Length: 34
	//                                 Target Name: cifs/192.168.71.7
	//                             Attribute: End of list
	//                                 NTLMV2 Response Item Type: End of list (0x0000)
	//                                 NTLMV2 Response Item Length: 0
	//                             padding: 00000000
	//                     Domain name: .
	//                         Length: 2
	//                         Maxlen: 2
	//                         Offset: 88
	//                     User name: auditor
	//                         Length: 14
	//                         Maxlen: 14
	//                         Offset: 90
	//                     Host name: WIN-LAIER32GF
	//                         Length: 26
	//                         Maxlen: 26
	//                         Offset: 104
	//                     Session Key: 91397c36277a9fce38c1cb83d7ee66dc
	//                         Length: 16
	//                         Maxlen: 16
	//                         Offset: 472
	//                      […]Negotiate Flags: 0xe2888215, Negotiate 56, Negotiate Key Exchange, Negotiate 128, Negotiate Version, Negotiate Target Info, Negotiate Extended Session Security, Negotiate Always Sign, Negotiate NTLM key, Negotiate Sign, Request Targe
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
	//                         .... .... .... ..0. .... .... .... .... = Target Type Server: Not set
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
	//                     Version 10.0 (Build 26100); NTLM Current Revision 15
	//                         Major Version: 10
	//                         Minor Version: 0
	//                         Build Number: 26100
	//                         NTLM Current Revision: 15
	//                     MIC: cb60bac4d825353456ed59130710c76c
	//                 mechListMIC: 01000000dbe7d2f9150ca89e00000000
	//                 NTLMSSP Verifier
	//                     Version Number: 1
	//                     Verifier Body: dbe7d2f9150ca89e00000000
	//

	packetReference := []byte{0x4e, 0x54, 0x4c, 0x4d, 0x53, 0x53, 0x50, 0x0, 0x2, 0x0, 0x0, 0x0, 0x1e, 0x0, 0x1e, 0x0, 0x38, 0x0, 0x0, 0x0, 0x15, 0x82, 0x8a, 0xe2, 0x2f, 0x9d, 0xaf, 0xe6, 0xd9, 0x95, 0xfa, 0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x98, 0x0, 0x98, 0x0, 0x56, 0x0, 0x0, 0x0, 0xa, 0x0, 0xf4, 0x65, 0x0, 0x0, 0x0, 0xf, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x2, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x1, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x4, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x3, 0x0, 0x1e, 0x0, 0x57, 0x0, 0x49, 0x0, 0x4e, 0x0, 0x2d, 0x0, 0x46, 0x0, 0x32, 0x0, 0x54, 0x0, 0x44, 0x0, 0x36, 0x0, 0x4c, 0x0, 0x55, 0x0, 0x54, 0x0, 0x30, 0x0, 0x50, 0x0, 0x52, 0x0, 0x7, 0x0, 0x8, 0x0, 0x99, 0x3b, 0xdf, 0x1, 0x1c, 0x83, 0xdb, 0x1, 0x0, 0x0, 0x0, 0x0}

	NTLMChallenge := NTLMSSP_CHALLENGE{}
	err := NTLMChallenge.Read(packetReference)
	if err != nil {
		t.Fatal(err)
	}

	equal, str := CompareBytesSlices(packetReference, NTLMChallenge.ToBytes())
	if equal != 0 {
		t.Fatal(str)
	}
}
