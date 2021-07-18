package gopsocks

import (
	"fmt"
	"io"
	"net"

	"github.com/hophouse/gop/utils"
)

func RunClientSocks() {
	// Socks listener
	addressSocks := "127.0.0.1:1337"
	fmt.Println("[+] Start a socks server on ", addressSocks)

	addressSocksTcp := net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 1337,
		Zone: "",
	}

	l, err := net.ListenTCP("tcp", &addressSocksTcp)
	if err != nil {
		utils.Log.Fatalln(err)
	}

	for {
		// Contact the tunnel
		addressTunnel := "127.0.0.1:1338"
		fmt.Println("[+] Establish connexion to the tunnel at ", addressTunnel)

		connTunnel, err := net.Dial("tcp", addressTunnel)
		if err != nil {
			utils.Log.Fatalln(err)
		}

		connSocks, err := l.AcceptTCP()
		if err != nil {
			utils.Log.Fatalln(err)
		}

		fmt.Println("[+] Socks server received a connexion from ", connSocks.RemoteAddr().String())

		go handleClientSocksConnexion(connTunnel, connSocks)
		connTunnel.Close()
	}
}

func RunServerSocks() {
	address := "127.0.0.1:1338"
	fmt.Println("[+] Start the server tunnel on ", address)

	addressTcp := net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 1338,
	}

	l, err := net.ListenTCP("tcp", &addressTcp)
	if err != nil {
		utils.Log.Fatalln(err)
	}

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			utils.Log.Fatalln(err)
		}

		fmt.Println("[+] Tunnel server received a connexion from ", conn.RemoteAddr().String())

		go handleServerTunnelConnexion(conn)
	}
}

func handleServerTunnelConnexion(conn net.Conn) {
	defer conn.Close()

	fmt.Println("[+] Handling server tunnel negociation")
	network, address, err := handleSocksServerNegociation(conn)
	if err != nil {
		fmt.Println("\t[!] ", err)
	}

	fmt.Println("[+] Connexion to the socks client established")
	newConn, _ := net.Dial(network, address)
	defer newConn.Close()

	go io.Copy(newConn, conn)
	io.Copy(conn, newConn)
}

func handleClientSocksConnexion(connTunnel net.Conn, connSocks net.Conn) {
	defer connTunnel.Close()
	defer connSocks.Close()

	fmt.Println("[+] Handling client negociation")
	err := handleSocksClientNegociation(connTunnel, connSocks)
	if err != nil {
		utils.Log.Panicln(err)
	}
	fmt.Println("[+] Connexion to the socks server established")

	go io.Copy(connSocks, connTunnel)
	io.Copy(connTunnel, connSocks)
}
