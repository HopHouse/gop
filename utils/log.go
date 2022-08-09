package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

//func init() {
//	NewLogger("log-f")
//}

func NewLogger(logFileName string) {
	var err error

	// Open Log File
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}

	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("Initialisation of the log file")
}

func CloseLogger(file *os.File) {
	file.Close()
}

// New logger for commands that do not need to create a complete directory structure
func NewLoggerStdout() {
	Log = log.New(os.Stdout, "", log.Lmsgprefix)
}

func NewLoggerNull() {
	Log = log.New(io.Discard, "", log.Lmsgprefix)
}

func NewLoggerStdoutDateTimeFile() {
	Log = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

func NewLoggerStdoutDateTime() {
	Log = log.New(os.Stdout, "", log.LstdFlags)
}

func CreateOutputDir(directoryName string, commandName string) {
	var err error

	//  Main directory structure of the ouput
	t := time.Now()
	if directoryName == "" {
		directoryName = t.Format("20060102-15-04-05") + "-gop-" + commandName
	}

	if _, err := os.Stat(directoryName); os.IsNotExist(err) {
		errDir := os.MkdirAll(directoryName, 0755)
		if errDir != nil {
			log.Fatal(errDir)
		}
	}

	directoryName = directoryName + "/"

	// Move to the new generated folder
	os.Chdir(directoryName)
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Printf("[-] Error when created log dir %s", mydir)
	}
}
