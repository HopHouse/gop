package utils

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

func ReadRequestFromFile(filename string) (*http.Request, error) {
	var httpRequestPtr *http.Request

	f, err := os.Open(filename)
	if err != nil {
		logger.Println(err)
		logger.Println(err)
		return nil, err
	}

	requestByte, err := io.ReadAll(f)
	if err != nil {
		logger.Println(err)
		logger.Println(err)
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
				logger.Println(err)
				logger.Println(err)
				return nil, err
			}
		} else {
			logger.Println(err)
			logger.Println(err)
			return nil, err
		}
	}

	return httpRequestPtr, nil
}
