package gopRelay

// func TestSMBNegotiateProtocoleRequest(t *testing.T) {

// 	netBIOSReference := []byte{0x0, 0x0, 0x1, 0x1c}

// 	smbHeaderReference := []byte{0xfe, 0x53, 0x4d, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

// 	// Initialize the encoder and decoder.  Normally enc and dec would be
// 	// bound to network connections and the encoder and decoder would
// 	// run in different processes.
// 	var buffer bytes.Buffer // Stand-in for a network connection
// 	buffer.Write(netBIOSReference)
// 	// buffer.Write(smbHeaderReference)

// 	// // SMB Header
// 	var smbHeader SMB2_HEADER_SYNC

// 	err = smbHeader.Read()
// 	if err != nil {
// 		t.Fatal("decode error:", err)
// 	}

// 	b := append(netBIOSPacket.ToBytes(), smbHeader.ToBytes()...)
// 	// b = append(b, smb2NegotiateRequest.ToBytes()...)

// 	compareStatus, compareStr := CompareBytesSlices(append(netBIOSReference, smbHeaderReference...), b)
// 	if compareStatus != 0 {
// 		errorStr := bytes.NewBuffer([]byte{})
// 		// fmt.Fprint(errorStr, q.ToString())
// 		fmt.Fprint(errorStr, compareStr)
// 		t.Fatal(errorStr.String())
// 	}
// }
