package goptunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/hophouse/gop/utils/logger"
)

// Taken from the RFC 1928
// From March 1996
// Link : https://datatracker.ietf.org/doc/html/rfc1928

type versionMethodMessageStruct struct {
	/*
		+----+----------+----------+
		|VER | NMETHODS | METHODS  |
		+----+----------+----------+
		| 1  |    1     | 1 to 255 |
		+----+----------+----------+
	*/
	ver      byte
	nmethods byte
	methods  []byte
}

func (versionMethodMessage *versionMethodMessageStruct) read(buffer []byte) error {
	/*
		if len(buffer) < 3 {
			errMsg := fmt.Sprintf("Request size %d is lower than the minimal expected size (3 bytes)", len(buffer))
			return errors.New(errMsg)
		}
		if len(buffer) > 257 {
			errMsg := fmt.Sprintf("Request size %d is greater than the maximal expected size (257 bytes)", len(buffer))
			return errors.New(errMsg)
		}
	*/

	versionMethodMessage.ver = buffer[0]
	versionMethodMessage.nmethods = buffer[1]
	methodsSize := int(versionMethodMessage.nmethods)
	/*
		if len(buffer) != 2+methodsSize {
			errMsg := fmt.Sprintf("Request size %d when compiled is not the one expected", len(buffer))
			return errors.New(errMsg)
		}
	*/

	versionMethodMessage.methods = buffer[2 : 2+methodsSize]

	return nil
}

type methodSelectionMessageStruct struct {
	/*
	   +----+--------+
	   |VER | METHOD |
	   +----+--------+
	   | 1  |   1    |
	   +----+--------+
	*/
	ver    byte
	method byte
}

func (methodSelectionMessage *methodSelectionMessageStruct) make() ([]byte, error) {
	methodSelectionMessage.ver = 0x05
	methodSelectionMessage.method = 0x00

	buffer := make([]byte, 2)

	buffer[0] = methodSelectionMessage.ver
	buffer[1] = methodSelectionMessage.method

	return buffer, nil
}

type requestStruct struct {
	/*
	   +----+-----+-------+------+----------+----------+
	   |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	   +----+-----+-------+------+----------+----------+
	   | 1  |  1  | X'00' |  1   | Variable |    2     |
	   +----+-----+-------+------+----------+----------+

	*/
	ver      byte
	cmd      byte
	rsv      byte
	atyp     byte
	dst_addr []byte
	dst_port []byte
}

func (request requestStruct) print(starter string) error {
	logger.Println(starter, "Request :")
	logger.Println(starter, "Version                 :", request.ver)
	cmdString := "Unknown"
	switch request.cmd {
	case 0x01:
		cmdString = "CONNECT"
	case 0x02:
		cmdString = "BIND"
	case 0x03:
		cmdString = "UDP ASSOCIATE"
	}
	logger.Printf("%s Command                 : %b (%s)\n", starter, request.cmd, cmdString)

	logger.Println(starter, "Reserved                :", request.rsv)
	atypString := "Unknown"
	switch request.atyp {
	case 0x01:
		atypString = "IPv4"
	case 0x03:
		atypString = "domain"
	case 0x04:
		atypString = "IPv6"
	}
	logger.Printf("%s Atyp                    : %v (%s)\n", starter, request.atyp, atypString)

	switch request.atyp {
	case 0x01:
		logger.Println(starter, "Destination address     :", request.dst_addr)
	case 0x03:
		logger.Println(starter, "Destination address     :", string(request.dst_addr))
	case 0x04:
		logger.Println(starter, "Destination address     :", request.dst_addr)
	}

	dst_port := binary.BigEndian.Uint16(request.dst_port)
	logger.Printf("%s Destination port        : %v (%d)\n", starter, request.dst_port, dst_port)

	return nil
}

func (request requestStruct) getNetwork() (string, error) {
	network := ""

	switch request.cmd {
	case 0x01:
		network = "tcp"
	case 0x02:
		network = "tcp"
	case 0x03:
		network = "udp"
	}

	return network, nil
}

func (request requestStruct) getAddress() (string, error) {
	address := ""
	switch request.atyp {
	case 0x01:
		ip := net.IPv4(request.dst_addr[0], request.dst_addr[1], request.dst_addr[2], request.dst_addr[3])
		address = ip.String()
	case 0x03:
		ips, err := net.LookupIP(string(request.dst_addr))
		if err != nil {
			return address, err
		}
		address = ips[0].String()
	case 0x04:
	default:
		return address, nil
	}

	dst_portUint16 := binary.BigEndian.Uint16(request.dst_port)
	dst_portString := strconv.FormatUint(uint64(dst_portUint16), 10)
	address = address + ":" + dst_portString

	return address, nil
}

// Return the error code corresponding to the REP field in the response
func (request requestStruct) testConnexion() byte {
	network, err := request.getNetwork()
	if err != nil {
		return 0x03
	}

	address, err := request.getAddress()
	if err != nil {
		return 0x09
	}

	logger.Printf("\t[+] Trying connection to %s %s\n", network, address)
	_, err = net.Dial(network, address)
	if err != nil {
		return 0x04
	}

	return 0x00
}

func (request *requestStruct) read(buffer []byte) error {
	// Minimal valid request size is 8 bytes
	if len(buffer) < 8 {
		errorMsg := fmt.Sprintf("Request size %d is lower than the minimal expected size (8 bytes)", len(buffer))
		return errors.New(errorMsg)
	}

	request.ver = buffer[0]
	request.cmd = buffer[1]
	request.rsv = buffer[2]
	request.atyp = buffer[3]

	var end int
	switch request.atyp {
	case 0x01:
		request.dst_addr = make([]byte, 4)
		end = 4 + 4
		request.dst_addr = buffer[4:end]
	case 0x03:
		size := int(buffer[4])
		request.dst_addr = make([]byte, size)
		end = 5 + size
		request.dst_addr = buffer[5:end]
	case 0x04:
		request.dst_addr = make([]byte, 16)
		end = 4 + 16
		request.dst_addr = buffer[4:end]
	default:
		errorMsg := fmt.Sprintf("Request destination address size %x is not valid", request.atyp)
		return errors.New(errorMsg)
	}

	request.dst_port = buffer[end : end+2]

	return nil
}

type responseStruct struct {
	/*
	   +----+-----+-------+------+----------+----------+
	   |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	   +----+-----+-------+------+----------+----------+
	   | 1  |  1  | X'00' |  1   | Variable |    2     |
	   +----+-----+-------+------+----------+----------+

	*/
	ver       byte
	rep       byte
	rsv       byte
	atyp      byte
	bind_addr []byte
	bind_port []byte
}

func (response *responseStruct) read(buffer []byte) error {
	end := len(buffer)

	response.ver = buffer[0]
	response.rep = buffer[1]
	response.rsv = buffer[2]
	response.atyp = buffer[3]
	response.bind_addr = append([]byte{}, buffer[5:end-2]...)
	response.bind_port = buffer[end-3 : end]

	return nil
}

func (response *responseStruct) make(request requestStruct, rep byte) {
	response.ver = 0x05
	response.rep = rep
	response.rsv = 0x00
	response.atyp = request.atyp
	response.bind_addr = request.dst_addr
	response.bind_port = request.dst_port
}

func (response *responseStruct) toBytes() ([]byte, error) {
	buffer := make([]byte, 4)

	buffer[0] = response.ver
	buffer[1] = response.rep
	buffer[2] = response.rsv
	buffer[3] = response.atyp
	buffer = append(buffer, response.bind_addr...)
	buffer = append(buffer, response.bind_port...)

	// Validate that response is well formed
	r := responseStruct{}
	err := r.read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func handleSocksServerNegociation(tunnel tunnelInterface) (string, string, error) {
	logger.Println("[+] Running socks proxy negociation phase.")
	versionMethodMessage := versionMethodMessageStruct{}
	methodSelectionMessage := methodSelectionMessageStruct{}
	request := requestStruct{}
	response := responseStruct{}

	buf := make([]byte, 1500)
	n, err := tunnel.Read(buf)
	if err != nil {
		return "", "", err
	}
	logger.Println("\t[+] Received version method with", n, "bytes.")

	err = versionMethodMessage.read(buf[:n])
	if err != nil {
		return "", "", err
	}

	// Send a METHOD selection message
	methodSelectionMessageBuff, err := methodSelectionMessage.make()
	if err != nil {
		return "", "", err
	}
	logger.Println("\t[+] Sending method selection")
	_, err = tunnel.Write(methodSelectionMessageBuff)
	if err != nil {
		logger.Printf("Error write : %s\n", err)
	}

	// Receive first request
	buf = make([]byte, 4096)
	n, err = tunnel.Read(buf)
	if err != nil {
		return "", "", err
	}
	logger.Println("\t[+] Received request with", n, "bytes.")
	logger.Println("[=]", buf[:n])

	err = request.read(buf[:n])
	if err != nil {
		return "", "", err
	}
	request.print("\t\t[+]")

	logger.Println("\t[+] Test request connexion.")
	returnCode := request.testConnexion()
	logger.Printf("\t\t[+] Connexion return code : %x\n", returnCode)

	logger.Println("\t[+] Make response.")
	response.make(request, returnCode)

	logger.Println("\t[+] Transform response to bytes.")
	responseBuff, err := response.toBytes()
	if err != nil {
		return "", "", err
	}
	logger.Println("\t[+] Send response with size :", len(responseBuff))
	_, err = tunnel.Write(responseBuff)
	if err != nil {
		logger.Printf("Error write : %s\n", err)
	}

	network, _ := request.getNetwork()
	address, _ := request.getAddress()
	return network, address, nil
}

//
// func handleSocksClientNegociation(connTunnel net.Conn, connSocks net.Conn) error {
// 	logger.Println("[+] Running socks proxy negociation phase.")
//
// 	// Receive Version Identifier from porxy client
// 	buf := make([]byte, 4096)
// 	n, err := connSocks.Read(buf)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Send Version Identifier to proxy server through the tunnel
// 	_, err = connTunnel.Write(buf[:n])
// 	if err != nil {
// 		logger.Printf("Error write : %s\n", err)
// 	}
//
// 	// Receive method selection message from the socks server trough tunnel
// 	buf = make([]byte, 4096)
// 	n, err = connTunnel.Read(buf)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Send method selection message to the proxy client
// 	_, err = connSocks.Write(buf[:n])
// 	if err != nil {
// 		logger.Printf("Error write : %s\n", err)
// 	}
//
// 	// Receive request from proxy client
// 	buf = make([]byte, 4096)
// 	n, err = connSocks.Read(buf)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Send request to sock server through tunnel
// 	_, err = connTunnel.Write(buf[:n])
// 	if err != nil {
// 		logger.Printf("Error write : %s\n", err)
// 	}
//
// 	// Receive response from sock server through tunnel
// 	buf = make([]byte, 4096)
// 	n, err = connTunnel.Read(buf)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Send response to the proxy client
// 	_, err = connSocks.Write(buf[:n])
// 	if err != nil {
// 		logger.Printf("Error write : %s\n", err)
// 	}
//
// 	return nil
// }
