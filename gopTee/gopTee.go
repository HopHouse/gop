package goptee

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hophouse/gop/utils/logger"
)

func RunTeeCmd(outputFile string, cmdOption string) error {
	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_APPEND, 0744)
	if err != nil {
		return err
	}

	file.WriteString(fmt.Sprintf("[%s]\n", time.Now().Format("2006-01-02 03:04:05")))
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	file.WriteString(path)
	cmdString := fmt.Sprintf("> %s\n", cmdOption)
	file.WriteString(cmdString)

	cmdStringSlicePipe := strings.Split(cmdOption, "|")
	logger.Println(cmdStringSlicePipe)

	stack := make([]*exec.Cmd, 0)

	for i, input := range cmdStringSlicePipe {
		var cmd *exec.Cmd

		logger.Printf("i: %d - %s\n", i, input)
		cmdStringSlice := strings.Split(strings.TrimSpace(input), " ")

		if len(cmdStringSlice) <= 1 {
			cmd = exec.Command(cmdStringSlice[0])
		} else {
			cmd = exec.Command(cmdStringSlice[0], cmdStringSlice[1:]...)
		}

		stack = append(stack, cmd)

	}

	err = ExecPipeCommands(stack...)
	if err != nil {
		return err
	}

	return nil
}

func ExecPipeCommands(stack ...*exec.Cmd) error {
	pipeSlice := make([]*io.PipeWriter, len(stack)-1)

	var out bytes.Buffer

	logger.Printf("Len stack : %d\n", len(stack))
	for i := 0; i < len(stack)-1; i++ {
		inPipe, outPipe := io.Pipe()
		logger.Printf("Pipe %d :\n\t%#v\n\t%#v\n", i, &inPipe, &outPipe)

		pipeSlice[i] = outPipe
		stack[i+1].Stdin = inPipe
		stack[i].Stdout = outPipe
		stack[i].Stderr = os.Stderr
	}
	stack[len(stack)-1].Stdout = &out

	logger.Println("[+] Start the Exec Stack")

	err := ExecStackCmd(stack, pipeSlice)
	if err != nil {
		return err
	}

	logger.Printf("[+] Final output:\n%s\n", out.String())

	return nil
}

func ExecStackCmd(stack []*exec.Cmd, pipeSlice []*io.PipeWriter) error {
	logger.Printf("Len stack : %d\n", len(stack))
	for i := (len(stack) - 1); i >= 0; i-- {
		logger.Printf("Starting Stack %d\n", i)
		stack[i].Start()
	}

	pipeSlice[0].Close()
	stack[1].Start()
	stack[1].Wait()
	logger.Println(stack[1].Output())

	/*
		if i == (len(cmdStringSlicePipe) - 1) {
			output, err := cmd.CombinedOutput()
			if err != nil {
				file.WriteString(err.Error())
				return err
			}

			file.Write(output)
			file.WriteString("\n")

			logger.Println(string(output))
		} else {
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatal(err)
			}
		}

	*/
	return nil
}
