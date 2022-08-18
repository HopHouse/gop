package goptunnel

import (
	"io"
	"net"

	"github.com/google/uuid"
	"github.com/hophouse/gop/utils"
)

func handleServerSend(currentTunnel tunnelInterface, socketAddress string, goRoutineUUID uuid.UUID) {
	// Contact the host that will receive traffic
	connSocket, err := net.Dial("tcp", socketAddress)
	if err != nil {
		utils.Log.Fatalln(err)
		return
	}
	utils.Log.Println(goRoutineUUID.String(), "Socket received traffic. Will send message")

	done := make(chan bool, 2)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		utils.Log.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
		io.Copy(currentTunnel, connSocket)
		done <- true
	}(currentTunnel, connSocket)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		utils.Log.Println(goRoutineUUID.String(), "Copy currentSocks<-connTunnel")
		io.Copy(connSocket, currentTunnel)
		done <- true
	}(currentTunnel, connSocket)

	<-done

	func() {
		utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
		connSocket.Close()
	}()

	func() {
		utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
		currentTunnel.Close()
	}()
}

func handleServerListen(currentTunnel tunnelInterface, socketListener net.Listener, goRoutineUUID uuid.UUID) {
	connSocket, err := socketListener.Accept()
	if err != nil {
		utils.Log.Println(err)
		return
	}
	utils.Log.Println(goRoutineUUID.String(), "Socket received traffic. Will send message")

	utils.Log.Println("Send \"send\" message")
	currentTunnel.Write([]byte("send"))
	utils.Log.Println("Sent \"send\" message")

	done := make(chan bool)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		utils.Log.Println(goRoutineUUID.String(), "Copy currentSocks<-connTunnel")
		_, err := io.Copy(connSocket, currentTunnel)
		if err != nil {
			utils.Log.Println(goRoutineUUID.String(), "Copy error :", err)
		}

		done <- true
	}(currentTunnel, connSocket)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		utils.Log.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
		_, err := io.Copy(currentTunnel, connSocket)
		if err != nil {
			utils.Log.Println(goRoutineUUID.String(), "Copy error :", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	<-done

	func() {
		utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
		currentTunnel.Close()
	}()

	func() {
		utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
		connSocket.Close()
	}()
}

func RunServer(tunnelAddress string, socketAddress string, tunnelType string, mode string) {
	var tunnel tunnelInterface

	switch tunnelType {
	case "plain":
		tunnel = &PlainTextTunnel{
			Protocol: "tcp",
			Address:  tunnelAddress,
			Conn:     nil,
			Listener: nil,
		}
		/*
			case "udp-plain":
				tunnel = &UDPPlainTextTunnel{
					Protocol: "udp",
					Address:  tunnelAddress,
					Conn:     nil,
					Listener: nil,
				}
		*/
	case "tls":
		tunnel = &TlsTunnel{
			Protocol: "tcp",
			Address:  tunnelAddress,
			Conn:     nil,
			Listener: nil,
		}
	case "http":
		tunnel = &HTTPPlainTextTunnel{
			Protocol: "tcp",
			Address:  tunnelAddress,
			Conn:     nil,
			Listener: nil,
		}
	default:
		utils.Log.Fatalln("Unknown type")
	}

	utils.Log.Println("Tunnel listen to :", tunnelAddress)

	// Start a listener for the tunnel
	err := tunnel.Listen()
	if err != nil {
		utils.Log.Fatalln(err)
	}

	switch mode {
	case "send":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			currentTunnel := tunnel
			err = currentTunnel.Accept()
			if err != nil {
				//utils.Log.Fatalln(err)
				utils.Log.Println(err)
				continue
			}
			utils.Log.Println(goRoutineUUID.String(), "Tunnel established with ", currentTunnel.RemoteAddr())

			go handleServerSend(currentTunnel, socketAddress, goRoutineUUID)

		}
	case "listen":

		// Start a listener for the socket
		socketListener, err := net.Listen("tcp", socketAddress)
		if err != nil {
			utils.Log.Fatalln(err)
		}
		utils.Log.Println("Local listen address to send traffic to is :", socketAddress)

		for {
			goRoutineUUID, _ := uuid.NewRandom()

			currentTunnel := tunnel.Clone()
			err := currentTunnel.Accept()
			if err != nil {
				utils.Log.Fatalln(err)
			}
			utils.Log.Println(goRoutineUUID.String(), "Tunnel established with ", currentTunnel.RemoteAddr())

			go handleServerListen(currentTunnel, socketListener, goRoutineUUID)
		}
	case "socks5":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			currentTunnel := tunnel
			err = currentTunnel.Accept()
			if err != nil {
				utils.Log.Fatalln(err)
			}
			utils.Log.Println(goRoutineUUID.String(), "Tunnel received a connexion from ", currentTunnel.RemoteAddr())

			go handleServerSocks5Connexion(currentTunnel)
		}
	default:
		utils.Log.Fatalln("Unknown mode")
	}

}

func handleClientSend(currentTunnel tunnelInterface, socketAddress string, goRoutineUUID uuid.UUID) {
	done := make(chan bool)
	utils.Log.Printf("%s %x %#v\n", goRoutineUUID.String(), &currentTunnel, currentTunnel)

	// Contact the host
	connSocket, err := net.Dial("tcp", socketAddress)
	if err != nil {
		utils.Log.Fatalln(err)
	}
	utils.Log.Println(goRoutineUUID.String(), "Establish connexion with client established at", socketAddress)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		utils.Log.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
		_, err := io.Copy(currentTunnel, connSocket)
		if err != nil {
			utils.Log.Println(goRoutineUUID.String(), "Copy error :", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		utils.Log.Println(goRoutineUUID.String(), "Copy connSocket<-currentTunnel")
		_, err := io.Copy(connSocket, currentTunnel)
		if err != nil {
			utils.Log.Println(goRoutineUUID.String(), "Copy error :", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	<-done

	func() {
		utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
		connSocket.Close()
	}()
	func() {
		utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
		currentTunnel.Close()
	}()
}

func RunClient(tunnelAddress string, socketAddress string, tunnelType string, mode string) {
	var tunnel tunnelInterface

	switch tunnelType {
	case "plain":
		tunnel = &PlainTextTunnel{
			Protocol: "tcp",
			Address:  tunnelAddress,
			Conn:     nil,
			Listener: nil,
		}
		/*
			case "udp-plain":
				tunnel = &UDPPlainTextTunnel{
					Protocol: "udp",
					Address:  tunnelAddress,
					Conn:     nil,
					Listener: nil,
				}
		*/
	case "tls":
		tunnel = &TlsTunnel{
			Protocol: "tcp",
			Address:  tunnelAddress,
			Conn:     nil,
			Listener: nil,
		}
	case "http":
		tunnel = &HTTPPlainTextTunnel{
			Protocol: "tcp",
			Address:  tunnelAddress,
			Conn:     nil,
			Listener: nil,
		}
	default:
		utils.Log.Fatalln("Unknown type")
	}

	switch mode {
	case "send":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			// need to be defined
			currentTunnel := tunnel.Clone()
			err := currentTunnel.Dial()
			if err != nil {
				utils.Log.Fatalln(err)
			}
			utils.Log.Println(goRoutineUUID.String(), "Tunnel connexion established with", tunnelAddress)

			for {
				tun := make([]byte, 150)
				n, _ := currentTunnel.Read(tun)

				if n > 0 {
					if string(tun[:n]) == "send" {
						break
					}
				}
			}
			utils.Log.Println(goRoutineUUID.String(), "\"send\" message received", currentTunnel.RemoteAddr())

			go handleClientSend(currentTunnel, socketAddress, goRoutineUUID)
		}
	case "listen":
		utils.Log.Println("Local listen address to send traffic is :", socketAddress)
		socketListen, err := net.Listen("tcp", socketAddress)
		if err != nil {
			utils.Log.Fatalln(err)
			return
		}

		for {
			goRoutineUUID, _ := uuid.NewRandom()

			connSocket, err := socketListen.Accept()
			if err != nil {
				utils.Log.Fatalln(err)
				continue
			}
			utils.Log.Println(goRoutineUUID.String(), "Establish connexion with", socketAddress)

			go func(connSocket net.Conn) {
				// Contact the tunnel
				currentTunnel := tunnel.Clone()
				err := currentTunnel.Dial()
				if err != nil {
					utils.Log.Fatalln(err)
					return
				}

				done := make(chan bool)

				defer func() {
					utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
					currentTunnel.Close()
				}()

				go func(currentTunnel tunnelInterface, connSocket net.Conn) {
					utils.Log.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
					io.Copy(currentTunnel, connSocket)
					done <- true
				}(currentTunnel, connSocket)

				go func(currentTunnel tunnelInterface, connSocket net.Conn) {
					utils.Log.Println(goRoutineUUID.String(), "Copy connSocket<-currentTunnel")
					io.Copy(connSocket, currentTunnel)
					done <- true
				}(currentTunnel, connSocket)

				<-done

				defer func() {
					utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
					connSocket.Close()
				}()
			}(connSocket)
		}
	case "socks5":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			// Contact the tunnel
			currentTunnel := tunnel.Clone()
			err := currentTunnel.Dial()
			if err != nil {
				utils.Log.Fatalln(err)
				continue
			}
			utils.Log.Println(goRoutineUUID.String(), "Tunnel connexion established with", tunnelAddress)

			defer func() {
				utils.Log.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
				currentTunnel.Close()
			}()

			for {
				tun := make([]byte, 150)
				n, err := currentTunnel.Read(tun)
				if err != nil {
					utils.Log.Println(err)
					continue
				}
				utils.Log.Println("Read content")

				if n > 0 {
					utils.Log.Printf("Received : %#v\n", n)
					if string(tun[:n]) == "send" {
						break
					}
				}
			}

			go handleServerSocks5Connexion(currentTunnel)
		}
	default:
		utils.Log.Fatalln("Unknown mode")
	}
}

func handleServerSocks5Connexion(tunnel tunnelInterface) {
	done := make(chan bool)

	utils.Log.Println("[+] Handling server tunnel negociation")
	network, address, err := handleSocksServerNegociation(tunnel)
	if err != nil {
		utils.Log.Println("\t[!]", err)
		return
	}
	utils.Log.Println("[+] Connexion to the socks client established", network, address)

	if network == "" || address == "" {
		utils.Log.Println("[!] Wrong info", network, address)
		return
	}
	newConn, err := net.Dial(network, address)
	if err != nil {
		utils.Log.Println(err)
		return
	}
	defer newConn.Close()

	go func() {
		utils.Log.Println("Run io.Copy currentSocks <- connTunnel")
		io.Copy(newConn, tunnel)
		done <- true
	}()

	go func() {
		utils.Log.Println("Run io.Copy currentTunnel <- connSocket")
		io.Copy(tunnel, newConn)
		done <- true
	}()

	<-done
}
