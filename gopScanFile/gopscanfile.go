package gopscanfile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hophouse/gop/utils/logger"
)

var (
	regList               []*regexp.Regexp
	locationBlackListPtr  *[]string
	extensionWhiteListPtr *[]string
	extensionBlackListPtr *[]string
	onlyFilesPtr          *bool
)

// RunSearchCmd : Run the Search command
func RunSearchCmd(patternList []string, pathList []string, locationBlackList []string, extensionWhiteList []string, extensionBlackList []string, onlyFiles bool, concurrency int) {
	inputChan := make(chan string)
	workerChan := make(chan bool, concurrency)
	locationBlackListPtr = &locationBlackList
	extensionWhiteListPtr = &extensionWhiteList
	extensionBlackListPtr = &extensionBlackList
	onlyFilesPtr = &onlyFiles

	// Compile all patterns
	for _, expr := range patternList {
		regExp, err := regexp.Compile(expr)
		if err != nil {
			logger.Printf("[X] Error with expression : %s\n", expr)
			break
		}
		regList = append(regList, regExp)
	}

	// Run workers
	for i := 0; i < concurrency; i++ {
		go func(inputChan chan string, workerChan chan bool) {
			for path := range inputChan {
				err := filepath.Walk(path, findInPath)
				if err != nil {
					logger.Printf("Error during walk in location : %s\n", path)
					logger.Println(err)
				}
			}
			workerChan <- true
		}(inputChan, workerChan)
	}

	// Walk from each given location in order to found files
	for _, path := range pathList {
		inputChan <- path
	}
	close(inputChan)

	// Wait for the workers to finish their jobs
	for i := 0; i < concurrency; i++ {
		<-workerChan
	}
	close(workerChan)
}

func findInPath(path string, info os.FileInfo, err error) error {
	// If there is an error, then do not run search routine
	if err != nil {
		return nil
	}

	// Check if we should avoid a location
	for _, location := range *locationBlackListPtr {
		if strings.HasPrefix(path, location) {
			return nil
		}
	}

	// Apply white list option. If extension file is blacklist then do to check the file
	if len(*extensionWhiteListPtr) > 0 {
		found := false
		for _, extension := range *extensionWhiteListPtr {
			if strings.HasSuffix(info.Name(), "."+extension) {
				found = true
			}
		}

		if !found {
			return nil
		}

	} else {
		// Apply black list option. If extension file is blacklist then do to check the file
		for _, extension := range *extensionBlackListPtr {
			extensionClean := strings.TrimSpace(extension)
			if strings.HasSuffix(info.Name(), "."+extensionClean) {
				return nil
			}
		}
	}

	for _, re := range regList {
		res := re.MatchString(info.Name())
		if res {
			if info.IsDir() && !*onlyFilesPtr {
				logger.Printf("[+] [D] %s\n", path)
			} else {
				logger.Printf("[+] [F] %s\n", path)
			}
			logger.Printf("     |_ Name : %s\n", info.Name())
			logger.Printf("     |_ Size : %d\n", info.Size())
			logger.Printf("     |_ Last modified : %s\n", info.ModTime())
			logger.Printf("     |_ Permissions : %s\n", info.Mode())
			logger.Printf("     |_ Matched Pattern : %s\n", re.String())
			logger.Printf("\n")
		}
	}

	// return errors.New("Could not find in the path.")
	return nil
}
