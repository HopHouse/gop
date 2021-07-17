package goirc

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hophouse/gop/utils"
	"github.com/jroimartin/gocui"
)

type optionsStruct struct {
	host     string
	port     string
	username string
}

var (
	Options     optionsStruct
	G           *gocui.Gui
	sendChan    chan string
	receiveChan chan string
)

func RunServerIRC(host string, port string, username string) {
	Options = optionsStruct{
		host:     host,
		port:     port,
		username: username,
	}

	sendChan = make(chan string)
	receiveChan = make(chan string)

	conn := launchServer(Options)
	defer conn.Close()

	go sendMessage(conn, sendChan)
	go receiveMessage(conn, receiveChan)
	go GuiReceiveMessage(receiveChan)

	mainGUI()
}

func RunClientIRC(host string, port string, username string) {
	Options = optionsStruct{
		host:     host,
		port:     port,
		username: username,
	}

	sendChan = make(chan string)
	receiveChan = make(chan string)
	defer close(sendChan)
	defer close(receiveChan)

	conn := connectToServer(Options)
	defer conn.Close()

	go sendMessage(conn, sendChan)
	go receiveMessage(conn, receiveChan)
	go GuiReceiveMessage(receiveChan)

	mainGUI()
}

func launchServer(options optionsStruct) net.Conn {
	address := fmt.Sprintf("%s:%s", options.host, options.port)
	fmt.Println("[+] Launching server on ", address)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", address)
	if err != nil {
		utils.Log.Panicln("Error listening:", err.Error())
	}

	conn, err := l.Accept()
	if err != nil {
		utils.Log.Panicln("Error accepting client:", err.Error())
	}

	return conn
}

func connectToServer(options optionsStruct) net.Conn {
	address := fmt.Sprintf("%s:%s", options.host, options.port)
	fmt.Println("[+] Connecting to ", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		utils.Log.Panicln(err)
	}

	return conn
}

func sendMessage(conn net.Conn, sendChan <-chan string) {
	for sendItem := range sendChan {
		conn.Write([]byte(sendItem))
		G.Update(func(g *gocui.Gui) error { return nil })
	}
}

func receiveMessage(conn net.Conn, receiveChan chan<- string) {
	for {
		message := ""
		for {
			buffer := make([]byte, 4096)
			n, _ := conn.Read(buffer)
			message += string(buffer[:n])
			if n < 4096 {
				break
			}
		}
		receiveChan <- message
		G.Update(func(g *gocui.Gui) error { return nil })
	}
}

func GuiReceiveMessage(receiveChan chan string) {
	for receiveItem := range receiveChan {
		message := receiveItem
		view, err := G.View("chat")
		if err == nil {
			view.Write([]byte(message))
			G.Update(func(g *gocui.Gui) error { return nil })
		}
	}
}

func executeCommand(command string) {
	if strings.HasPrefix(strings.ToLower(command), "username") {
		if len(command) < len("username")+1 {
			message := "[!] No username specified\n"
			receiveChan <- message
			return
		}

		newUsername := strings.Split(strings.SplitAfter(command, "username ")[1], "\n")[0]

		message := fmt.Sprintf("[+] %s changed username for : %s\n", Options.username, newUsername)
		sendChan <- message
		receiveChan <- message

		Options.username = newUsername
		view, _ := G.View("username")
		view.Clear()
		fmt.Fprint(view, Options.username+" > ")
		G.Update(func(g *gocui.Gui) error { return nil })

		return
	} else if strings.HasPrefix(strings.ToLower(command), "quit") {
		message := fmt.Sprintf("[+] %s leaves the chat\n", Options.username)
		sendChan <- message
		receiveChan <- message

		close(receiveChan)
		close(sendChan)

		G.Close()
		os.Exit(0)

	} else {
		receiveChan <- getHelp()
	}

}

func getHelp() string {
	helpMessage := []string{"",
		"[+] Help menu :",
		"\t!help : Displays help message",
		"\t!username : Changes username",
		"\t!quit : Leaves the chat",
		"",
		"",
	}

	return strings.Join(helpMessage, "\n")
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("chat", 0, 0, maxX, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Autoscroll = true
		v.Title = "Chat"
	}

	if v, err := g.SetView("username", 0, maxY-2, len(Options.username)+4, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(v, Options.username)
		fmt.Fprint(v, " > ")
	}

	if v, err := g.SetView("input", len(Options.username)+5, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
		v.Editable = true
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func GuiSendMessage(g *gocui.Gui, v *gocui.View) error {
	message := v.Buffer()
	if len(message) > 0 {
		if strings.HasPrefix(message, "!") {
			executeCommand(message[1:])
		} else {
			if !strings.HasPrefix(message, "[+] ") {
				message = fmt.Sprintf("%s > %s", Options.username, message)
			}
			sendChan <- message
			receiveChan <- message
		}
		v.Clear()
		v.SetCursor(0, 0)
	}

	return nil
}

func mainGUI() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		utils.Log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.SetCurrentView("input")
	g.Cursor = false

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, GuiSendMessage); err != nil {
		utils.Log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		utils.Log.Panicln(err)
	}

	// Make object global
	G = g

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		utils.Log.Panicln(err)
	}
}
