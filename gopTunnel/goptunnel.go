package goptunnel

import (
	"fmt"
	"io"
	"net"

	"github.com/hophouse/gop/utils"
)

func RunServer(tunnelAddress string, socketAddress string, mode string) {
	fmt.Println("[+] Tunnel listen to :", tunnelAddress)

	// Start a listener for the tunnel
	tunnelListner, err := net.Listen("tcp", tunnelAddress)
	if err != nil {
		utils.Log.Fatalln(err)
	}

	switch mode {
	case "send":
		fmt.Println("[+] Traffic will be redirected to :", socketAddress)

		for {
			connTunnel, err := tunnelListner.Accept()
			if err != nil {
				//utils.Log.Fatalln(err)
				utils.Log.Println(err)
				continue
			}
			fmt.Println("[+] Tunnel established with ", connTunnel.RemoteAddr().String())

			go func() {
				// Contact the host that will receive traffic
				connHost, err := net.Dial("tcp", socketAddress)
				if err != nil {
					//utils.Log.Fatalln(err)
					utils.Log.Println(err)
					return
				}

				go io.Copy(connHost, connTunnel)
				io.Copy(connTunnel, connHost)

				connHost.Close()
				connTunnel.Close()
			}()
		}
	case "listen":
		// Start a listener for the socket
		socketListener, err := net.Listen("tcp", socketAddress)
		if err != nil {
			utils.Log.Fatalln(err)
		}
		fmt.Println("[+] Local listen address to send traffic is :", socketAddress)

		for {
			connTunnel, err := tunnelListner.Accept()
			if err != nil {
				utils.Log.Fatalln(err)
				//utils.Log.Println(err)
			}
			fmt.Println("[+] Tunnel established with ", connTunnel.RemoteAddr().String())

			go func() {
				connListner, err := socketListener.Accept()
				if err != nil {
					utils.Log.Println(err)
					return
				}
				fmt.Println("[+] Socket received traffic. Will send message")

				connTunnel.Write([]byte("send"))

				go io.Copy(connTunnel, connListner)
				io.Copy(connListner, connTunnel)

				connListner.Close()
				connTunnel.Close()
			}()
		}

	case "socks5":
		for {
			connTunnel, err := tunnelListner.Accept()
			if err != nil {
				utils.Log.Fatalln(err)
			}
			fmt.Println("[+] Tunnel received a connexion from ", connTunnel.RemoteAddr().String())

			go handleServerSocks5Connexion(connTunnel)
		}
	default:
		utils.Log.Fatalln("Unknown mode")
	}

}

func RunClient(tunnelAddress string, socketAddress string, mode string) {
	switch mode {
	case "send":
		for {
			// Contact the tunnel
			connTunnel, err := net.Dial("tcp", tunnelAddress)
			if err != nil {
				utils.Log.Fatalln(err)
			}
			fmt.Println("[+] Tunnel connexion established with", tunnelAddress)

			// Contact the host
			connSocks, err := net.Dial("tcp", socketAddress)
			if err != nil {
				utils.Log.Fatalln(err)
			}
			fmt.Println("[+] Establish connexion with client established with", socketAddress)

			for {
				tun := make([]byte, 4096)
				n, _ := connTunnel.Read(tun)

				if n > 0 {
					if string(tun[:n]) == "send" {
						break
					}
				}
			}

			go func() {

				go io.Copy(connSocks, connTunnel)
				io.Copy(connTunnel, connSocks)

				connSocks.Close()
				connTunnel.Close()
			}()
		}
	case "listen":
		fmt.Println("[+] Local listen address to send traffic is :", socketAddress)
		socketListen, err := net.Listen("tcp", socketAddress)
		if err != nil {
			utils.Log.Fatalln(err)
		}

		for {
			connListner, err := socketListen.Accept()
			if err != nil {
				utils.Log.Fatalln(err)
			}
			fmt.Println("[+] Establish connexion with", socketAddress)

			go func() {
				// Contact the tunnel
				connTunnel, err := net.Dial("tcp", tunnelAddress)
				if err != nil {
					utils.Log.Fatalln(err)
				}
				fmt.Println("[+] Tunnel connexion established with", tunnelAddress)

				go io.Copy(connListner, connTunnel)
				io.Copy(connTunnel, connListner)

				connTunnel.Close()
			}()
		}
	case "socks5":
		for {
			// Contact the tunnel
			connTunnel, err := net.Dial("tcp", tunnelAddress)
			if err != nil {
				utils.Log.Fatalln(err)
			}
			fmt.Println("[+] Tunnel connexion established with", tunnelAddress)

			for {
				tun := make([]byte, 4096)
				n, _ := connTunnel.Read(tun)

				if n > 0 {
					if string(tun[:n]) == "send" {
						break
					}
				}
			}

			go handleServerSocks5Connexion(connTunnel)
		}
	default:
		utils.Log.Fatalln("Unknown mode")
	}

}

func handleServerSocks5Connexion(conn net.Conn) {
	defer conn.Close()

	fmt.Println("[+] Handling server tunnel negociation")
	network, address, err := handleSocksServerNegociation(conn)
	if err != nil {
		fmt.Println("\t[!] ", err)
	}
	fmt.Println("[+] Connexion to the socks client established", network, address)

	if network == "" || address == "" {
		utils.Log.Println("[!] Wrong info", network, address)
		return
	}
	newConn, err := net.Dial(network, address)
	if err != nil {
		//utils.Log.Panicln(err)
		return
	}
	defer newConn.Close()

	go io.Copy(newConn, conn)
	io.Copy(conn, newConn)
}
