package notification

import (
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/hophouse/gop/utils"
)

type NotifierStruct struct {
	OperatingSystem string
	User            string
}

var Notifier NotifierStruct

func init() {
	username := ""
	currentUser, err := user.Current()
	if err == nil {
		if runtime.GOOS == "windows" {
			username = strings.Split(currentUser.Username, "\\")[1]
		} else {
			username = currentUser.Username
		}
	}

	Notifier = NotifierStruct{
		OperatingSystem: runtime.GOOS,
		User:            username,
	}
}

func NotifyAndWait(notification string) {
	if Notifier.OperatingSystem == "windows" {
		cmd := exec.Command("C:\\Windows\\System32\\msg.exe", Notifier.User, "/W", notification)
		err := cmd.Start()
		if err != nil {
			utils.Log.Println(err)
		}
		cmd.Wait()
	}
}

func Notify(notification string) {
	if Notifier.OperatingSystem == "windows" {
		err := exec.Command("C:\\Windows\\System32\\msg.exe", Notifier.User, notification).Start()
		if err != nil {
			utils.Log.Println(err)
		}
	}
}
