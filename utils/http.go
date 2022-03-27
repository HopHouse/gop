package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func ReadRequestFromFile(filename string) (*http.Request, error) {
	httpRequestPtr := &http.Request{}

	f, err := os.Open(filename)
	if err != nil {
		Log.Println(err)
		fmt.Println(err)
		return nil, err
	}

	requestByte, err := io.ReadAll(f)
	if err != nil {
		Log.Println(err)
		fmt.Println(err)
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewReader(requestByte))

	httpRequestPtr, err = http.ReadRequest(reader)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			requestString := string(requestByte) + "\n\n"
			requestReader := bufio.NewReader(strings.NewReader(requestString))

			httpRequestPtr, err = http.ReadRequest(requestReader)
			if err != nil {
				Log.Println(err)
				fmt.Println(err)
				return nil, err
			}
		} else {
			Log.Println(err)
			fmt.Println(err)
			return nil, err
		}
	}

	return httpRequestPtr, nil
}
