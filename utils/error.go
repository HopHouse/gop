package utils

import "github.com/hophouse/gop/utils/logger"

func CheckError(err error) bool {
	if err != nil {
		logger.Println(err)
		return true
	}
	return false
}

func CheckErrorExit(err error) bool {
	if err != nil {
		logger.Fatal(err)
		return true
	}
	return false
}
