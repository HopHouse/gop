package gopStation

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
)

type agentStruct struct {
	Uuid      uuid.UUID
	startTime time.Time
	name      string
	address   string
	kind      string
	commands  []commandStruct
	conn      net.Conn
}

type commandStruct struct {
	input  string
	output string
}

func RunServerCmd(tcpAddress string, sslAddress string) {
	agentList := map[uuid.UUID]*agentStruct{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
	}()

	/*
	 * TCP
	 */
	if tcpAddress != "" {
		logger.Println("[+] Binding TCP listener to ", tcpAddress)

		listenerTcp, err := net.Listen("tcp", tcpAddress)
		if err != nil {
			log.Fatal("[!] Unable to bind the address")
		}
		defer listenerTcp.Close()

		go func(listener net.Listener) {
			for {
				agentConn, err := listener.Accept()
				if err != nil {
					logger.Println(err)
				}

				agentUuid := uuid.New()
				agentDate := time.Now()

				agentList[agentUuid] = &agentStruct{
					name:      "Undefined",
					address:   agentConn.RemoteAddr().String(),
					startTime: agentDate,
					Uuid:      agentUuid,
					conn:      agentConn,
					kind:      "TCP",
				}
				logger.Printf("\n[+] Accepting TCP connection from %s with UUID %s\n", agentConn.RemoteAddr().String(), agentUuid.String())
			}
		}(listenerTcp)
	}

	/*
	 * SSL
	 */
	if sslAddress != "" {
		logger.Println("[+] Binding SSL listener to ", sslAddress)

		serverCert, serverKey := gopproxy.GenerateCA()

		caBytes, err := x509.CreateCertificate(rand.Reader, serverCert, serverCert, serverKey.Public(), serverKey)
		if err != nil {
			logger.Fatal(err)
		}

		serverCertPEM := new(bytes.Buffer)
		err = pem.Encode(serverCertPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		})
		if err != nil {
			logger.Println(err)
		}

		serverPrivKeyPEM := new(bytes.Buffer)
		err = pem.Encode(serverPrivKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
		})
		if err != nil {
			logger.Println(err)
		}

		cer, err := tls.X509KeyPair(serverCertPEM.Bytes(), serverPrivKeyPEM.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}

		// Listen for incoming connections.
		listenerSsl, err := tls.Listen("tcp", sslAddress, config)
		if err != nil {
			log.Fatal("[!] Unable to bind the SSL address")
		}
		defer listenerSsl.Close()

		go func(listener net.Listener) {
			for {
				agentConn, err := listener.Accept()
				if err != nil {
					logger.Println(err)
				}

				agentUuid := uuid.New()
				agentDate := time.Now()

				agentList[agentUuid] = &agentStruct{
					name:      "Undefined",
					address:   agentConn.RemoteAddr().String(),
					startTime: agentDate,
					Uuid:      agentUuid,
					conn:      agentConn,
					kind:      "SSL",
				}
				logger.Printf("\n[+] Accepting SSL connection from %s with UUID %s\n", agentConn.RemoteAddr().String(), agentUuid.String())
			}
		}(listenerSsl)
	}

	/*
	 * Agent handling
	 */
	var currentAgent *agentStruct = nil

	reader := bufio.NewReader(os.Stdin)

	for {
		if currentAgent != nil {
			logger.Printf("%s > ", currentAgent.Uuid.String())
		} else {
			logger.Printf("$> ")
		}

		input, _ := reader.ReadString('\n')
		// convert CRLF to LF
		input = strings.ReplaceAll(input, "\n", "")
		commands := strings.Split(strings.TrimSpace(input), " ")

		if currentAgent != nil {
			switch commands[0] {
			case "help":
				displayHelp()
				continue
			case "shell":
				runShell(currentAgent)
				continue
			case "back":
				currentAgent = nil
				continue
			case "stop":
				currentAgent.conn.Close()
				delete(agentList, currentAgent.Uuid)
				currentAgent = nil
				continue
			case "name":
				if len(commands) < 2 {
					logger.Print("\n[!] Please provide a name\n")
					continue
				}
				currentAgent.name = strings.Join(commands[1:], " ")
				continue
			case "info":
				w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)
				logger.Fprintf(w, "UUID :\t%s\n", currentAgent.Uuid.String())
				logger.Fprintf(w, "Name :\t%s\n", currentAgent.name)
				logger.Fprintf(w, "Address :\t%s\n", currentAgent.address)
				logger.Fprintf(w, "Date :\t%s\n", currentAgent.startTime.Format("2006-01-02 03:04:05"))
				logger.Fprintf(w, "Kind :\t%s\n", currentAgent.kind)
				w.Flush()
				continue
			case "history":
				if len(commands) == 1 {
					w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)
					logger.Fprint(w, "#\tCommand\n")
					for i, elem := range currentAgent.commands {
						logger.Fprintf(w, "%d\t%s\n", i, elem.input)
					}
					w.Flush()
				} else {
					targetCmd, err := strconv.Atoi(commands[1])
					if err != nil {
						logger.Print("\n[!] Please provide a valid number\n")
						continue
					}
					if targetCmd > len(currentAgent.commands) || targetCmd < 0 {
						logger.Print("\n[!] Please provide a valid number\n")
						continue
					}
					cmd := currentAgent.commands[targetCmd]
					logger.Printf("> %s\n%s\n", cmd.input, cmd.output)
				}
				continue
			default:
				continue
			}
		} else {
			switch commands[0] {
			case "help":
				displayHelp()
			case "list":
				w := tabwriter.NewWriter(os.Stdout, 36, 2, 2, ' ', 0)
				logger.Fprint(w, "UUID\tName\tKind\tAddress\tStart date\n")
				for _, agent := range agentList {
					logger.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", agent.Uuid.String(),
						agent.name, agent.kind, agent.address,
						agent.startTime.Format("2006-01-02 03:04:05"))
				}
				w.Flush()
				continue
			case "use":
				if len(commands) < 2 {
					logger.Print("\n[!] Please provide a UUID\n")
					continue
				}
				targetUuid, err := uuid.Parse(commands[1])
				if err != nil {
					logger.Printf("\n[!] Error parsing UUID %s\n", targetUuid)
					continue
				}

				agent, present := agentList[targetUuid]
				if !present {
					logger.Printf("\n[!] Could not find UUID %s\n", targetUuid)
					continue
				}
				currentAgent = agent
				continue
			case "name":
				if len(commands) < 2 {
					logger.Print("\n[!] Please provide a UUID\n")
					continue
				}
				if len(commands) < 3 {
					logger.Print("\n[!] Please provide a new name\n")
					continue
				}
				targetUuid, err := uuid.Parse(commands[1])
				if err != nil {
					logger.Printf("\n[!] Error parsing UUID %s\n", targetUuid)
					continue
				}

				agent, present := agentList[targetUuid]
				if !present {
					logger.Printf("\n[!] Could not find UUID %s\n", targetUuid)
					continue
				}
				agent.name = strings.Join(commands[2:], " ")
				continue
			case "stop":
				if len(commands) < 2 {
					logger.Print("\n[!] Please provide a UUID\n")
					continue
				}
				targetUuid, err := uuid.Parse(commands[1])
				if err != nil {
					logger.Printf("\n[!] Error parsing UUID %s\n", targetUuid)
					continue
				}

				agent, present := agentList[targetUuid]
				if !present {
					logger.Printf("\n[!] Could not find UUID %s\n", targetUuid)
					continue
				}
				agent.conn.Close()
				delete(agentList, agent.Uuid)
				currentAgent = nil
				continue
			case "exit":
				for _, agent := range agentList {
					agent.conn.Close()
				}
				os.Exit(0)
				continue
			default:
				continue
			}
		}
	}
}

func displayHelp() {
	logger.Printf("[+] Help from main\n")
	logger.Printf("\t- help\tDisplay this help\n")
	logger.Printf("\t- list\tGet a list of all the agents\n")
	logger.Printf("\t- use [agent]\tSelect an agent based on its UUID\n")
	logger.Printf("\t- stop [agent]\tStop specified agent by its UUID\n")
	logger.Printf("\t- name [agent] [name]\tRename specified agent by its UUID\n")
	logger.Printf("\t- exit\texit\n")

	logger.Printf("\n[+] Help from an agent\n")
	logger.Printf("\t- help\tDisplay this help\n")
	logger.Printf("\t- shell\tRun a shell inside\n")
	logger.Printf("\t- back\treturn to main menu from the agent\n")
	logger.Printf("\t- stop\tStop the agent and close the connection\n")
	logger.Printf("\t- name [name]\tChange name of the agent\n")
	logger.Printf("\t- history\tDisplay all the command executed into the agent\n")
	logger.Printf("\t- history [id]\tDisplay input and output of the command \"id\"\n")
	logger.Printf("\t- info\tGet information about the current agent\n")

	logger.Printf("\n[+] Help from an agent shell\n")
	logger.Printf("\t- !help\tDisplay this help\n")
	logger.Printf("\t- !back\treturn to the agent\n")
}

func runShell(currentAgent *agentStruct) {
	go func() {
		_, err := io.Copy(os.Stdout, currentAgent.conn)
		if err != nil {
			logger.Println(err)
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		// convert CRLF to LF
		input = strings.ReplaceAll(input, "\n", "")
		command := strings.TrimSpace(input)
		commandSlice := strings.Split(command, " ")

		switch commandSlice[0] {
		case "!help":
			displayHelp()
			continue
		case "exit":
			logger.Println("[?] If you want to exit, please use !back and close the connection with the stop command.")
			continue
		case "!back":
			return
		default:
			if command == "" {
				continue
			}

			_, err := io.WriteString(currentAgent.conn, command+"\n")
			if err != nil {
				logger.Println(err)
				return
			}

			currentAgent.commands = append(currentAgent.commands, commandStruct{
				input:  command,
				output: "",
			})

			// io.Copy(conn, command)
			continue
		}
	}
}
