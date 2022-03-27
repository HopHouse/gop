package gopwebtampering

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/hophouse/gop/utils"
)

func TamperHostHeader(webRequestFilename string) error {
	originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	if err != nil {
		utils.Log.Fatalln(err)
	}
	defer originalRequest.Body.Close()

	originalBody, err := io.ReadAll(originalRequest.Body)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		utils.Log.Fatalln(err)
	}
	originalRequest.Body = ioutil.NopCloser(bytes.NewReader(originalBody))

	originalRequest.RequestURI = ""

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		utils.Log.Fatalln(err)
	}

	fmt.Fprint(w, resultHeaderToString(maxLenHeader, maxLenLocalhostAddresses))
	utils.Log.Print("[+] Ref:\n")
	utils.Log.Printf("- Protocol: %s \n", refHTTPProto)
	utils.Log.Printf("- Status code: %d \n", refHTTPStatusCode)
	utils.Log.Printf("- Content length: %d \n", refHTTPContentLength)

	fmt.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Reference", "", refHTTPStatusCode, refHTTPContentLength, ""))
	utils.Log.Print("[+] Ref:\n")
	fmt.Fprint(w, "\n")
	w.Flush()

	// Modify the Host header value
	utils.Log.Print("\n[+] Host header modification\n")
	for _, destination := range LocalhostAddresses {
		customRequest := originalRequest.Clone(originalRequest.Context())
		customRequest.URL, _ = url.Parse(originalRequest.URL.String())
		customRequest.Body = ioutil.NopCloser(bytes.NewReader(originalBody))

		customRequest.Host = destination

		proto, statusCode, contentLength, body, err := sendRequest(customRequest)
		if err != nil {
			fmt.Printf("[%s: %s] ERROR\n", "Host", destination)
			continue
		}
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		utils.Log.Printf("[%s: %s] %s %d %d %s\n", "Host", destination, proto, statusCode, contentLength, message)
		if message != "" {
			fmt.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Host", destination, statusCode, contentLength, message))
			utils.Log.Println(body)
			w.Flush()
		}
	}

	w.Flush()

	return nil
}

func TamperReferrerHeader(webRequestFilename string) error {
	err := doRequestAndAnalysis(webRequestFilename, func(request *http.Request, value string) {
		request.Header.Set("Referer", value)
	})
	if err != nil {
		utils.Log.Println(err)
		return err
	}

	return nil
}

type transformation func(*http.Request, string)

func doRequestAndAnalysis(webRequestFilename string, transformationFunc transformation) error {
	originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	if err != nil {
		utils.Log.Fatalln(err)
	}
	defer originalRequest.Body.Close()

	originalBody, err := io.ReadAll(originalRequest.Body)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		utils.Log.Fatalln(err)
	}
	originalRequest.Body = ioutil.NopCloser(bytes.NewReader(originalBody))

	originalRequest.RequestURI = ""

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		utils.Log.Fatalln(err)
	}

	fmt.Fprint(w, resultHeaderToString(maxLenHeader, maxLenLocalhostAddresses))
	utils.Log.Print("[+] Ref:\n")
	utils.Log.Printf("- Protocol: %s \n", refHTTPProto)
	utils.Log.Printf("- Status code: %d \n", refHTTPStatusCode)
	utils.Log.Printf("- Content length: %d \n", refHTTPContentLength)

	fmt.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Reference", "", refHTTPStatusCode, refHTTPContentLength, ""))
	utils.Log.Print("[+] Ref:\n")
	fmt.Fprint(w, "\n")
	w.Flush()

	utils.Log.Print("\n[+] Referer header modification\n")
	for _, destination := range LocalhostAddresses {
		customRequest := originalRequest.Clone(originalRequest.Context())
		customRequest.Body = ioutil.NopCloser(bytes.NewReader(originalBody))
		refererHeader := ""

		if customRequest.Referer() != "" {
			refererUrl, err := url.Parse(customRequest.Referer())
			if err != nil {
				fmt.Printf("[%s: %s] ERROR\n", "Referer", destination)
				continue
			}
			refererUrl.Host = destination
			refererHeader = refererUrl.String()
		} else {
			// Needed to create a copy
			refererUrl, _ := url.Parse(customRequest.URL.String())

			refererUrl.Host = destination
			refererHeader = refererUrl.String()
		}

		// Apply transformation
		transformationFunc(customRequest, refererHeader)

		proto, statusCode, contentLength, body, err := sendRequest(customRequest)
		if err != nil {
			fmt.Printf("[%s: %s] ERROR\n", "Referer", destination)
			continue
		}
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		utils.Log.Printf("[%s: %s] %s %d %d %s\n", "Referer", destination, proto, statusCode, contentLength, message)

		if message != "" {
			fmt.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Referer", destination, statusCode, contentLength, message))
			utils.Log.Println(body)
			w.Flush()
		}
	}

	w.Flush()

	return nil
}

func TamperIPSource(webRequestFilename string) error {
	originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	if err != nil {
		utils.Log.Fatalln(err)
	}
	defer originalRequest.Body.Close()

	originalBody, err := io.ReadAll(originalRequest.Body)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		utils.Log.Fatalln(err)
	}
	originalRequest.Body = ioutil.NopCloser(bytes.NewReader(originalBody))

	originalRequest.RequestURI = ""

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		utils.Log.Fatalln(err)
	}

	fmt.Fprint(w, resultHeaderToString(maxLenHeader, maxLenLocalhostAddresses))
	utils.Log.Print("[+] Ref:\n")
	utils.Log.Printf("- Protocol: %s \n", refHTTPProto)
	utils.Log.Printf("- Status code: %d \n", refHTTPStatusCode)
	utils.Log.Printf("- Content length: %d \n", refHTTPContentLength)

	fmt.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Reference", "", refHTTPStatusCode, refHTTPContentLength, ""))
	utils.Log.Print("[+] Ref:\n")
	fmt.Fprint(w, "\n")
	w.Flush()

	// Add headers
	utils.Log.Print("\n[+] Adding headers\n")
	for _, header := range HeadersIP {
		for _, destination := range LocalhostAddresses {
			customRequest := originalRequest.Clone(originalRequest.Context())
			customRequest.Body = ioutil.NopCloser(bytes.NewReader(originalBody))

			customRequest.Header.Del(header)
			customRequest.Header.Add(header, destination)

			proto, statusCode, contentLength, body, err := sendRequest(customRequest)
			if err != nil {
				fmt.Printf("[%s: %s] ERROR\n", header, destination)
				continue
			}
			message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
			utils.Log.Printf("[%s: %s] %s %d %d %s\n", header, destination, proto, statusCode, contentLength, message)

			if message != "" {
				fmt.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, header, destination, statusCode, contentLength, message))
				utils.Log.Println(body)
				w.Flush()

			}
		}
	}

	w.Flush()

	return nil
}

func sendRequest(request *http.Request) (proto string, statusCode int, contentLength int, body string, err error) {
	var proxyUrlPtr *url.URL = nil

	if Options.Proxy != "" {
		proxyUrlPtr, err = url.Parse(Options.Proxy)
		if err != nil {
			// Fatal because if reference value failed, then it is not possible to compare others correctly
			fmt.Println(err)
			utils.Log.Fatalln(err)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrlPtr),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 1 * time.Minute,
	}

	// First request to get information about the page and set reference values
	httpResponse, err := client.Do(request)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		fmt.Println(err)
		utils.Log.Fatalln(err)
	}
	defer httpResponse.Body.Close()

	bodyByte, _ := io.ReadAll(httpResponse.Body)
	body = string(bodyByte)
	contentLength = len(body)

	proto = httpResponse.Proto
	statusCode = httpResponse.StatusCode

	response, _ := httputil.DumpResponse(httpResponse, false)
	contentLength = contentLength + len(response)

	// bodyScanner := bufio.NewScanner(httpResponse.Body)
	// for bodyScanner.Scan() {
	// 	body = body + bodyScanner.Text()
	// }
	return proto, statusCode, contentLength, body, nil
}

func analyseDifferences(refProto string, refStatusCode int, refContentLength int, proto string, statusCode int, contentLength int) string {
	message := ""

	if refProto != proto {
		message += "Protocol is different. "
	}

	if refStatusCode != statusCode {
		message += "Status code is different. "
	}

	if refContentLength != contentLength {
		message += "Content length is different. "
	}

	return message
}

func resultToString(maxLenHeader int, maxLenValue int, header string, value string, statusCode int, contentLength int, comment string) string {
	for i := 0; utf8.RuneCountInString(header) < maxLenHeader; i++ {
		header = header + " "
	}

	for i := 0; utf8.RuneCountInString(value) < maxLenValue; i++ {
		value = value + " "
	}

	return fmt.Sprintf("%s\t%s\t%d\t%d\t%s\n", header, value, statusCode, contentLength, comment)
}

func resultHeaderToString(maxLenHeader int, maxLenValue int) string {
	header := "Name"
	value := "Value"

	for i := 0; utf8.RuneCountInString(header) < maxLenHeader; i++ {
		header = header + " "
	}

	for i := 0; utf8.RuneCountInString(value) < maxLenValue; i++ {
		value = value + " "
	}

	return fmt.Sprintf("%s\t%s\tStatus Code\tContent Length\tComment\n", header, value)
}
