package goptunnel

import (
	"io"
	"net"

	"github.com/google/uuid"
	"github.com/hophouse/gop/utils/logger"
)

func handleServerSend(currentTunnel tunnelInterface, socketAddress string, goRoutineUUID uuid.UUID) {
	// Contact the host that will receive traffic
	connSocket, err := net.Dial("tcp", socketAddress)
	if err != nil {
		logger.Fatalln(err)
		return
	}
	logger.Println(goRoutineUUID.String(), "Socket received traffic. Will send message")

	done := make(chan bool, 2)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		logger.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
		_, err := io.Copy(currentTunnel, connSocket)
		if err != nil {
			logger.Printf("Error during io.Copy : %s\n", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		logger.Println(goRoutineUUID.String(), "Copy currentSocks<-connTunnel")
		_, err := io.Copy(connSocket, currentTunnel)
		if err != nil {
			logger.Printf("Error during io.Copy : %s\n", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	<-done

	func() {
		logger.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
		connSocket.Close()
	}()

	func() {
		logger.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
		currentTunnel.Close()
	}()
}

func handleServerListen(currentTunnel tunnelInterface, socketListener net.Listener, goRoutineUUID uuid.UUID) {
	connSocket, err := socketListener.Accept()
	if err != nil {
		logger.Println(err)
		return
	}
	logger.Println(goRoutineUUID.String(), "Socket received traffic. Will send message")

	logger.Println("Send \"send\" message")
	_, err = currentTunnel.Write([]byte("send"))
	if err != nil {
		logger.Printf("Error during currentTunnel.Write : %s\n", err)
	}
	logger.Println("Sent \"send\" message")

	done := make(chan bool)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		logger.Println(goRoutineUUID.String(), "Copy currentSocks<-connTunnel")
		_, err := io.Copy(connSocket, currentTunnel)
		if err != nil {
			logger.Println(goRoutineUUID.String(), "Copy error :", err)
		}

		done <- true
	}(currentTunnel, connSocket)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		logger.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
		_, err := io.Copy(currentTunnel, connSocket)
		if err != nil {
			logger.Println(goRoutineUUID.String(), "Copy error :", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	<-done

	func() {
		logger.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
		currentTunnel.Close()
	}()

	func() {
		logger.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
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
		logger.Fatalln("Unknown type")
	}

	logger.Println("Tunnel listen to :", tunnelAddress)

	// Start a listener for the tunnel
	err := tunnel.Listen()
	if err != nil {
		logger.Fatalln(err)
	}

	switch mode {
	case "send":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			currentTunnel := tunnel
			err = currentTunnel.Accept()
			if err != nil {
				// logger.Fatalln(err)
				logger.Println(err)
				continue
			}
			logger.Println(goRoutineUUID.String(), "Tunnel established with ", currentTunnel.RemoteAddr())

			go handleServerSend(currentTunnel, socketAddress, goRoutineUUID)

		}
	case "listen":

		// Start a listener for the socket
		socketListener, err := net.Listen("tcp", socketAddress)
		if err != nil {
			logger.Fatalln(err)
		}
		logger.Println("Local listen address to send traffic to is :", socketAddress)

		for {
			goRoutineUUID, _ := uuid.NewRandom()

			currentTunnel := tunnel.Clone()
			err := currentTunnel.Accept()
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Println(goRoutineUUID.String(), "Tunnel established with ", currentTunnel.RemoteAddr())

			go handleServerListen(currentTunnel, socketListener, goRoutineUUID)
		}
	case "socks5":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			currentTunnel := tunnel
			err = currentTunnel.Accept()
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Println(goRoutineUUID.String(), "Tunnel received a connexion from ", currentTunnel.RemoteAddr())

			go handleServerSocks5Connexion(currentTunnel)
		}
	default:
		logger.Fatalln("Unknown mode")
	}
}

func handleClientSend(currentTunnel tunnelInterface, socketAddress string, goRoutineUUID uuid.UUID) {
	done := make(chan bool)
	logger.Printf("%s %x %#v\n", goRoutineUUID.String(), &currentTunnel, currentTunnel)

	// Contact the host
	connSocket, err := net.Dial("tcp", socketAddress)
	if err != nil {
		logger.Fatalln(err)
	}
	logger.Println(goRoutineUUID.String(), "Establish connexion with client established at", socketAddress)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		logger.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
		_, err := io.Copy(currentTunnel, connSocket)
		if err != nil {
			logger.Println(goRoutineUUID.String(), "Copy error :", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	go func(currentTunnel tunnelInterface, connSocket net.Conn) {
		logger.Println(goRoutineUUID.String(), "Copy connSocket<-currentTunnel")
		_, err := io.Copy(connSocket, currentTunnel)
		if err != nil {
			logger.Println(goRoutineUUID.String(), "Copy error :", err)
		}
		done <- true
	}(currentTunnel, connSocket)

	<-done

	func() {
		logger.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
		connSocket.Close()
	}()
	func() {
		logger.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
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
		logger.Fatalln("Unknown type")
	}

	switch mode {
	case "send":
		for {
			goRoutineUUID, _ := uuid.NewRandom()

			// need to be defined
			currentTunnel := tunnel.Clone()
			err := currentTunnel.Dial()
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Println(goRoutineUUID.String(), "Tunnel connexion established with", tunnelAddress)

			for {
				tun := make([]byte, 150)
				n, _ := currentTunnel.Read(tun)

				if n > 0 {
					if string(tun[:n]) == "send" {
						break
					}
				}
			}
			logger.Println(goRoutineUUID.String(), "\"send\" message received", currentTunnel.RemoteAddr())

			go handleClientSend(currentTunnel, socketAddress, goRoutineUUID)
		}
	case "listen":
		logger.Println("Local listen address to send traffic is :", socketAddress)
		socketListen, err := net.Listen("tcp", socketAddress)
		if err != nil {
			logger.Fatalln(err)
			return
		}

		for {
			goRoutineUUID, _ := uuid.NewRandom()

			connSocket, err := socketListen.Accept()
			if err != nil {
				logger.Fatalln(err)
				continue
			}
			logger.Println(goRoutineUUID.String(), "Establish connexion with", socketAddress)

			go func(connSocket net.Conn) {
				// Contact the tunnel
				currentTunnel := tunnel.Clone()
				err := currentTunnel.Dial()
				if err != nil {
					logger.Fatalln(err)
					return
				}

				done := make(chan bool)

				defer func() {
					logger.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
					currentTunnel.Close()
				}()

				go func(currentTunnel tunnelInterface, connSocket net.Conn) {
					logger.Println(goRoutineUUID.String(), "Copy currentTunnel<-connSocket")
					_, err := io.Copy(currentTunnel, connSocket)
					if err != nil {
						logger.Println(goRoutineUUID.String(), "Copy error :", err)
					}
					done <- true
				}(currentTunnel, connSocket)

				go func(currentTunnel tunnelInterface, connSocket net.Conn) {
					logger.Println(goRoutineUUID.String(), "Copy connSocket<-currentTunnel")
					_, err := io.Copy(connSocket, currentTunnel)
					if err != nil {
						logger.Println(goRoutineUUID.String(), "Copy error :", err)
					}
					done <- true
				}(currentTunnel, connSocket)

				<-done

				defer func() {
					logger.Println(goRoutineUUID.String(), "Close the connection in ", connSocket.RemoteAddr())
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
				logger.Fatalln(err)
				continue
			}
			logger.Println(goRoutineUUID.String(), "Tunnel connexion established with", tunnelAddress)

			defer func() {
				logger.Println(goRoutineUUID.String(), "Close the connection in ", currentTunnel.RemoteAddr())
				currentTunnel.Close()
			}()

			for {
				tun := make([]byte, 150)
				n, err := currentTunnel.Read(tun)
				if err != nil {
					logger.Println(err)
					continue
				}
				logger.Println("Read content")

				if n > 0 {
					logger.Printf("Received : %#v\n", n)
					if string(tun[:n]) == "send" {
						break
					}
				}
			}

			go handleServerSocks5Connexion(currentTunnel)
		}
	default:
		logger.Fatalln("Unknown mode")
	}
}

func handleServerSocks5Connexion(tunnel tunnelInterface) {
	done := make(chan bool)

	logger.Println("[+] Handling server tunnel negociation")
	network, address, err := handleSocksServerNegociation(tunnel)
	if err != nil {
		logger.Println("\t[!]", err)
		return
	}
	logger.Println("[+] Connexion to the socks client established", network, address)

	if network == "" || address == "" {
		logger.Println("[!] Wrong info", network, address)
		return
	}
	newConn, err := net.Dial(network, address)
	if err != nil {
		logger.Println(err)
		return
	}
	defer newConn.Close()

	go func() {
		logger.Println("Run io.Copy currentSocks <- connTunnel")
		_, err := io.Copy(newConn, tunnel)
		if err != nil {
			logger.Printf("Copy error : %s\n", err)
		}
		done <- true
	}()

	go func() {
		logger.Println("Run io.Copy currentTunnel <- connSocket")
		_, err := io.Copy(tunnel, newConn)
		if err != nil {
			logger.Printf("Copy error : %s\n", err)
		}
		done <- true
	}()

	<-done
}
