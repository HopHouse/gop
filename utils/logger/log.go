package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogType log.Logger

var Log *LogType
var mu sync.Mutex

func init() {
	NewLoggerStdoutDateTime()
}

func New(out io.Writer, prefix string, flag int) *LogType {
	return (*LogType)(log.New(out, prefix, flag))
}

func NewLogger(logFileName string) {
	var err error

	// Open Log File
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}

	Log = New(file, "", log.Lmsgprefix)
}

func NewLoggerDateTime(logFileName string) {
	var err error

	// Open Log File
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}

	Log = New(file, "", log.LstdFlags|log.Lmsgprefix)
}

func NewLoggerDateTimeFile(logFileName string) {
	var err error

	// Open Log File
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}

	Log = New(file, "", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
}

func CloseLogger(file *os.File) {
	file.Close()
}

// New logger for commands that do not need to create a complete directory structure
func NewLoggerStdout() {
	Log = New(os.Stdout, "", log.Lmsgprefix)
}

func NewLoggerNull() {
	Log = New(io.Discard, "", log.Lmsgprefix)
}

func NewLoggerStdoutDateTime() {
	Log = New(os.Stdout, "", log.LstdFlags|log.Lmsgprefix)
}

func NewLoggerStdoutDateTimeFile() {
	Log = New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
}

func CreateOutputDir(directoryName string, commandName string) string {
	//  Main directory structure of the ouput
	t := time.Now()

	if directoryName == "" {
		directoryName = t.Format("20060102-15-04-05") + "-gop-" + commandName
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	directoryName = filepath.Join(currentDirectory, directoryName)

	directoryName, _ = filepath.Abs(directoryName)
	if err != nil {
		log.Fatal(err)
	}
	directoryName = filepath.Clean(directoryName)

	if _, err := os.Stat(directoryName); os.IsNotExist(err) {
		errDir := os.MkdirAll(directoryName, 0755)
		if errDir != nil {
			log.Fatal(errDir)
		}
	}

	return directoryName
}

func Writer() io.Writer {
	log := (*log.Logger)(Log)
	return log.Writer()
}

func Print(v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Print(v...)
	} else {
		fmt.Print(v...)
		log.Print(v...)
	}
	mu.Unlock()
}

func Printf(format string, v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Printf(format, v...)
	} else {
		fmt.Printf(format, v...)
		log.Printf(format, v...)
	}
	mu.Unlock()
}

func Println(v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Println(v...)
	} else {
		fmt.Println(v...)
		log.Println(v...)
	}
	mu.Unlock()
}

func Fprint(w io.Writer, v ...any) (n int, err error) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		n, err = fmt.Fprint(w, v...)
	} else {
		n, err = fmt.Fprint(w, v...)
		n, err = fmt.Fprint(log.Writer(), v...)
	}

	mu.Unlock()
	return
}

func Fprintln(w io.Writer, v ...any) (n int, err error) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		n, err = fmt.Fprintln(w, v...)
	} else {
		n, err = fmt.Fprintln(w, v...)
		n, err = fmt.Fprintln(log.Writer(), v...)
	}
	mu.Unlock()

	return
}

func Fprintf(w io.Writer, format string, v ...any) (n int, err error) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		n, err = fmt.Fprintf(w, format, v...)
	} else {
		n, err = fmt.Fprintf(w, format, v...)
		n, err = fmt.Fprintf(log.Writer(), format, v...)
	}

	mu.Unlock()
	return
}

func Fatal(v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Fatal(v...)
	} else {
		fmt.Print(v...)
		log.Fatal(v...)
	}
	mu.Unlock()
}

func Fatalln(v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Fatalln(v...)
	} else {
		fmt.Println(v...)
		log.Fatalln(v...)
	}
	mu.Unlock()
}

func Fatalf(format string, v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Fatalf(format, v...)
	} else {
		fmt.Printf(format, v...)
		log.Fatalf(format, v...)
	}
	mu.Unlock()
}

func Panicln(v ...any) {
	mu.Lock()
	log := (*log.Logger)(Log)
	if log.Writer() == io.Discard {
		return
	}

	if log.Writer() == os.Stdout {
		log.Panicln(v...)
	} else {
		fmt.Println(v...)
		log.Panicln(v...)
	}

	mu.Unlock()
}
