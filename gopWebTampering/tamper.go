package gopwebtampering

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
	"golang.org/x/exp/slices"
)

func TamperHostHeader(webRequestFilename string) error {
	originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	if err != nil {
		logger.Println(err)
		return err
	}
	defer originalRequest.Body.Close()

	originalBody, err := io.ReadAll(originalRequest.Body)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}
	originalRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

	originalRequest.RequestURI = ""

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}

	logger.Fprint(w, resultHeaderToString(maxLenHeader, maxLenLocalhostAddresses))
	logger.Print("[+] Ref:\n")
	logger.Printf("- Protocol: %s \n", refHTTPProto)
	logger.Printf("- Status code: %d \n", refHTTPStatusCode)
	logger.Printf("- Content length: %d \n", refHTTPContentLength)

	logger.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Reference", "", refHTTPStatusCode, refHTTPContentLength, ""))
	logger.Print("[+] Ref:\n")
	logger.Fprint(w, "\n")
	w.Flush()

	// Modify the Host header value
	logger.Print("\n[+] Host header modification\n")
	for _, destination := range LocalhostAddresses {
		customRequest := originalRequest.Clone(originalRequest.Context())
		customRequest.URL, _ = url.Parse(originalRequest.URL.String())
		customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

		customRequest.Host = destination

		proto, statusCode, contentLength, body, err := sendRequest(customRequest)
		if err != nil {
			logger.Printf("[%s: %s] ERROR\n", "Host", destination)
			continue
		}
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		logger.Printf("[%s: %s] %s %d %d %s\n", "Host", destination, proto, statusCode, contentLength, message)
		if message != "" {
			logger.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Host", destination, statusCode, contentLength, message))
			logger.Println(body)
			w.Flush()
		}
	}

	w.Flush()

	return nil
}

func TamperReferrerHeader(webRequestFilename string) error {
	title := "Referer header modification"

	originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	if err != nil {
		logger.Println(err)
		return err
	}
	defer originalRequest.Body.Close()

	originalBody, err := io.ReadAll(originalRequest.Body)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}
	originalRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

	err = doRequestAndAnalysisTamperReferrerHeader(title, originalRequest, originalBody, func(request *http.Request, value string) {
		request.Header.Set("Referer", value)
	})

	if err != nil {
		logger.Println(err)
		return err
	}

	return nil
}

type transformation func(*http.Request, string)

func doRequestAndAnalysisTamperReferrerHeader(title string, originalRequest *http.Request, originalBody []byte, transformationFunc transformation) error {
	// originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	// if err != nil {
	// 	logger.Fatalln(err)
	// }
	// defer originalRequest.Body.Close()

	// originalBody, err := io.ReadAll(originalRequest.Body)
	// if err != nil {
	// 	// Fatal because if reference value failed, then it is not possible to compare others correctly
	// 	logger.Fatalln(err)
	// }
	// originalRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

	originalRequest.RequestURI = ""

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}

	logger.Fprint(w, resultHeaderToString(maxLenHeader, maxLenLocalhostAddresses))
	logger.Print("[+] Ref:\n")
	logger.Printf("- Protocol: %s \n", refHTTPProto)
	logger.Printf("- Status code: %d \n", refHTTPStatusCode)
	logger.Printf("- Content length: %d \n", refHTTPContentLength)

	logger.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Reference", "", refHTTPStatusCode, refHTTPContentLength, ""))
	logger.Print("[+] Ref:\n")
	logger.Fprint(w, "\n")
	w.Flush()

	logger.Printf("\n[+] %s\n", title)
	for _, destination := range LocalhostAddresses {
		customRequest := originalRequest.Clone(originalRequest.Context())
		customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))
		refererHeader := ""

		if customRequest.Referer() != "" {
			refererUrl, err := url.Parse(customRequest.Referer())
			if err != nil {
				logger.Printf("[%s: %s] ERROR\n", "Referer", destination)
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
			logger.Printf("[%s: %s] ERROR\n", "Referer", destination)
			continue
		}
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		logger.Printf("[%s: %s] %s %d %d %s\n", "Referer", destination, proto, statusCode, contentLength, message)

		if message != "" {
			logger.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Referer", destination, statusCode, contentLength, message))
			logger.Println(body)
			w.Flush()
		}
	}

	w.Flush()

	return nil
}

func TamperIPSource(webRequestFilename string) error {
	originalRequest, err := utils.ReadRequestFromFile(webRequestFilename)
	if err != nil {
		logger.Println(err)
		return err
	}
	defer originalRequest.Body.Close()

	originalBody, err := io.ReadAll(originalRequest.Body)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}
	originalRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

	originalRequest.RequestURI = ""

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}

	logger.Fprint(w, resultHeaderToString(maxLenHeader, maxLenLocalhostAddresses))
	logger.Print("[+] Ref:\n")
	logger.Printf("- Protocol: %s \n", refHTTPProto)
	logger.Printf("- Status code: %d \n", refHTTPStatusCode)
	logger.Printf("- Content length: %d \n", refHTTPContentLength)

	logger.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, "Reference", "", refHTTPStatusCode, refHTTPContentLength, ""))
	logger.Print("[+] Ref:\n")
	logger.Fprint(w, "\n")
	w.Flush()

	// Add headers
	logger.Print("\n[+] Adding headers\n")
	for _, header := range HeadersIP {
		for _, destination := range LocalhostAddresses {
			customRequest := originalRequest.Clone(originalRequest.Context())
			customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

			customRequest.Header.Del(header)
			customRequest.Header.Add(header, destination)

			proto, statusCode, contentLength, body, err := sendRequest(customRequest)
			if err != nil {
				logger.Printf("[%s: %s] ERROR\n", header, destination)
				continue
			}
			message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
			logger.Printf("[%s: %s] %s %d %d %s\n", header, destination, proto, statusCode, contentLength, message)

			if message != "" {
				logger.Fprint(w, resultToString(maxLenHeader, maxLenLocalhostAddresses, header, destination, statusCode, contentLength, message))
				logger.Println(body)
				w.Flush()

			}
		}
	}

	w.Flush()

	return nil
}

func NginxOffBySlash(webRequestFilename string, urlOption string, validResources []string, shownStatusCode []int) error {
	/*
	 * Taken from an Orange Tsai presentataion at blackhat.
	 *
	 * Strategy :
	 * 	- Remove last part of the URL
	 * 	- Try to move up to 1 level (/assets/)
	 *  - Try to access the resource without the (/assets)
	 *  - Try to access a resource without the / and a valid resource located at one top level (/assets/../settings.py) or (/assets/../static/app.js)
	 *  - Try to access a resource without the / (/assets../)
	 *  - Try to access a resource without the / and a valid resource (/assets../settings.py) or (/assets../static/app.js)
	 */
	var originalRequest *http.Request
	var originalBody []byte
	var err error

	if webRequestFilename != "" {
		originalRequest, err = utils.ReadRequestFromFile(webRequestFilename)
		if err != nil {
			logger.Println(err)
			return err
		}
		defer originalRequest.Body.Close()

		originalBody, err = io.ReadAll(originalRequest.Body)
		if err != nil {
			// Fatal because if reference value failed, then it is not possible to compare others correctly
			logger.Println(err)
			return err
		}
		originalRequest.Body = io.NopCloser(bytes.NewReader(originalBody))

		originalRequest.RequestURI = ""
	} else {
		requestUrl, err := url.Parse(urlOption)
		if err != nil {
			// Fatal because if reference value failed, then it is not possible to compare others correctly
			logger.Println(err)
			return err
		}
		originalRequest = &http.Request{
			URL: requestUrl,
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 16, 2, 2, ' ', 0)

	// First request to get information about the page and set reference values
	refHTTPProto, refHTTPStatusCode, refHTTPContentLength, _, err := sendRequest(originalRequest)
	if err != nil {
		// Fatal because if reference value failed, then it is not possible to compare others correctly
		logger.Println(err)
		return err
	}

	logger.Fprint(w, "Status Code\tContent Length\tURL\tComment\n")
	logger.Print("[+] Ref:\n")
	logger.Printf("- Protocol: %s \n", refHTTPProto)
	logger.Printf("- Status code: %d \n", refHTTPStatusCode)
	logger.Printf("- Content length: %d \n", refHTTPContentLength)

	logger.Fprintf(w, "%d\t%d\t%s\t%s\n", refHTTPStatusCode, refHTTPContentLength, originalRequest.URL.String(), "")
	logger.Print("[+] Ref:\n")
	logger.Fprint(w, "\n")
	w.Flush()

	// Add headers
	logger.Print("\n[+] Nginx off-by-slash\n")

	// Retrieve URL
	originalUrl := originalRequest.URL
	originalPath := originalUrl.EscapedPath()
	originalPathSlice := strings.Split(originalPath, "/")

	var (
		newPath       string
		newUrl        *url.URL
		customRequest *http.Request
		proto         string
		statusCode    int
		contentLength int
	)

	// Try to move up to 1 level and keep trailing slash
	newPath = strings.Join(originalPathSlice[:len(originalPathSlice)-1], "/") + "/"
	newUrl = &url.URL{
		Scheme: originalUrl.Scheme,
		Host:   originalUrl.Host,
		Path:   newPath,
	}

	customRequest = originalRequest.Clone(originalRequest.Context())
	customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))
	customRequest.URL = newUrl

	proto, statusCode, contentLength, _, err = sendRequest(customRequest)
	if err != nil {
		logger.Printf("[URL: %s] ERROR\n", newUrl.String())
	} else {
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		logger.Printf("[URL: %s] %s %d %d %s\n", newUrl.String(), proto, statusCode, contentLength, message)

		logger.Fprintf(w, "%d\t%d\t%s\t%s\n", statusCode, contentLength, customRequest.URL.String(), message)
	}

	w.Flush()

	// Try to access the resource without the /
	newPath = strings.Join(originalPathSlice[:len(originalPathSlice)-1], "/")
	newUrl = &url.URL{
		Scheme: originalUrl.Scheme,
		Host:   originalUrl.Host,
		Path:   newPath,
	}

	customRequest = originalRequest.Clone(originalRequest.Context())
	customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))
	customRequest.URL = newUrl

	proto, statusCode, contentLength, _, err = sendRequest(customRequest)
	if err != nil {
		logger.Printf("[URL: %s] ERROR\n", newUrl.String())
	} else {
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		logger.Printf("[URL: %s] %s %d %d %s\n", newUrl.String(), proto, statusCode, contentLength, message)

		logger.Fprintf(w, "%d\t%d\t%s\t%s\n", statusCode, contentLength, customRequest.URL.String(), message)
	}

	w.Flush()

	// Try to access a resource without the / and a valid resource located at one top level (/assets/../settings.py) or (/assets/../static/app.js)
	// newPath = strings.Join(originalPathSlice[:len(originalPathSlice)-1], "/") + "/../" + validResource
	// newUrl = &url.URL{
	// 	Scheme: originalUrl.Scheme,
	// 	Host:   originalUrl.Host,
	// 	Path:   newPath,
	// }

	// customRequest = originalRequest.Clone(originalRequest.Context())
	// customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))
	// customRequest.URL = newUrl

	// proto, statusCode, contentLength, body, err = sendRequest(customRequest)
	// if err != nil {
	// 	logger.Printf("[URL: %s] ERROR\n", newUrl.String())
	// } else {
	// 	message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
	// 	logger.Printf("[URL: %s] %s %d %d %s\n", newUrl.String(), proto, statusCode, contentLength, message)

	// 	logger.Fprintf(w, "%d\t%d\t%s\t%s\n", statusCode, contentLength, customRequest.URL.String(), message)
	// }

	// w.Flush()

	// Try to access a resource without the / (/assets../)
	newPath = strings.Join(originalPathSlice[:len(originalPathSlice)-1], "/") + "../"
	newUrl = &url.URL{
		Scheme: originalUrl.Scheme,
		Host:   originalUrl.Host,
		Path:   newPath,
	}

	customRequest = originalRequest.Clone(originalRequest.Context())
	customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))
	customRequest.URL = newUrl

	proto, statusCode, contentLength, _, err = sendRequest(customRequest)
	if err != nil {
		logger.Printf("[URL: %s] ERROR\n", newUrl.String())
	} else {
		message := analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
		logger.Printf("[URL: %s] %s %d %d %s\n", newUrl.String(), proto, statusCode, contentLength, message)

		logger.Fprintf(w, "%d\t%d\t%s\t%s\n", statusCode, contentLength, customRequest.URL.String(), message)
	}

	w.Flush()

	// Try to access a resource without the / and a valid resource (/assets../settings.py) or (/assets../static/app.js)
	for _, validResource := range validResources {
		newPath = strings.Join(originalPathSlice[:len(originalPathSlice)-1], "/") + "../" + validResource
		newUrl = &url.URL{
			Scheme: originalUrl.Scheme,
			Host:   originalUrl.Host,
			Path:   newPath,
		}

		customRequest = originalRequest.Clone(originalRequest.Context())
		customRequest.Body = io.NopCloser(bytes.NewReader(originalBody))
		customRequest.URL = newUrl

		proto, statusCode, contentLength, _, err = sendRequest(customRequest)
		if err != nil {
			logger.Printf("[URL: %s] ERROR\n", newUrl.String())
		} else {
			if len(shownStatusCode) <= 0 || slices.Contains(shownStatusCode, statusCode) {
				message := ""
				if statusCode == 200 {
					message = "The server might be vulnerable."
				} else {
					message = analyseDifferences(refHTTPProto, refHTTPStatusCode, refHTTPContentLength, proto, statusCode, contentLength)
				}
				logger.Printf("[URL: %s] %s %d %d %s\n", newUrl.String(), proto, statusCode, contentLength, message)

				logger.Fprintf(w, "%d\t%d\t%s\t%s\n", statusCode, contentLength, customRequest.URL.String(), message)
			}
		}

		w.Flush()
	}

	return nil
}

func sendRequest(request *http.Request) (proto string, statusCode int, contentLength int, body string, err error) {
	var proxyUrlPtr *url.URL = nil

	if Options.Proxy != "" {
		proxyUrlPtr, err = url.Parse(Options.Proxy)
		if err != nil {
			// Fatal because if reference value failed, then it is not possible to compare others correctly
			logger.Println(err)
			return "", 0, 0, "", err
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
		logger.Println(err)
		return "", 0, 0, "", err
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
