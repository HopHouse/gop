package gopRelay

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"slices"
	"testing"
)

func TestNetBIOSSetLength(t *testing.T) {
	// Details
	// Message Type : Session message (0x00)
	// Length: 170
	lengthReference := []byte{0x0, 0x0, 0xaa}

	n := NetBiosPacket{
		MessageType: NETBIOS_SESSION_MESSAGE,
	}
	n.SetLength(170)

	if slices.Compare(lengthReference, []byte(n.Length[:])) != 0 {
		t.Fatalf("Setting length to 170, does not works. Expected %v and got %v", lengthReference, n.Length)
	}
}

func TestNetBIOStoBytes(t *testing.T) {
	// Details
	// Message Type : Session message (0x00)
	// Length: 170
	netBIOSPacketReference := []byte{0x0, 0x0, 0x0, 0xaa}

	n := NetBiosPacket{
		MessageType: NETBIOS_SESSION_MESSAGE,
	}
	n.SetLength(170)

	if slices.Compare(netBIOSPacketReference, n.ToBytes()) != 0 {
		t.Fatalf("ToBytes() does not work. Expected %v and got %v", netBIOSPacketReference, n.ToBytes())
	}
}

func TestNetBIOSGobPackerEncDec(t *testing.T) {
	// Details
	// Message Type : Session message (0x00)
	// Length: 170
	netBIOSPacketReference := []byte{0x0, 0x0, 0x0, 0xaa}

	n := NetBiosPacket{
		MessageType: NETBIOS_SESSION_MESSAGE,
	}
	n.SetLength(170)

	// Initialize the encoder and decoder.  Normally enc and dec would be
	// bound to network connections and the encoder and decoder would
	// run in different processes.
	var buffer bytes.Buffer // Stand-in for a network connection

	enc := gob.NewEncoder(&buffer) // Will write to network.

	dec := gob.NewDecoder(&buffer) // Will read from network.

	// Encode (send) the value.
	err := enc.Encode(n)
	if err != nil {
		t.Fatal("encode error:", err)
	}

	// Decode (receive) the value.
	var q NetBiosPacket
	err = dec.Decode(&q)
	if err != nil {
		t.Fatal("decode error:", err)
	}

	compareStatus, compareStr := CompareBytesSlices(n.ToBytes(), q.ToBytes())
	if compareStatus != 0 {
		errorStr := bytes.NewBuffer([]byte{})
		fmt.Fprint(errorStr, q.ToString())
		fmt.Fprint(errorStr, compareStr)
		t.Fatal(errorStr.String())
	}

	compareStatus, compareStr = CompareBytesSlices(netBIOSPacketReference, q.ToBytes())
	if compareStatus != 0 {
		errorStr := bytes.NewBuffer([]byte{})
		fmt.Fprint(errorStr, q.ToString())
		fmt.Fprint(errorStr, compareStr)
		t.Fatal(errorStr.String())
	}
}
