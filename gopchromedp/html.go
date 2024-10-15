package gopchromedp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/hophouse/gop/utils/logger"
)

// TakeScreenShot Take screenshot of the pages.
func GetHTMLCode(item *Item, url string, directory string, proxy string, cookie string, timeout int) {
	// take screenshot for all urls
	if strings.HasSuffix(url, ".pdf") {
		logger.Fprintln(logger.Writer(), "[+] Do not take a the HTML content of the PDF ", url)
		return
	}

	options := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"),
	)

	// Init chrome context
	if proxy != "" {
		proxyAllocatorOption := chromedp.ProxyServer(proxy)
		options = append(options, proxyAllocatorOption)
	}

	ctxBase, ctxcancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer ctxcancel()

	ccontext, ccancel := chromedp.NewContext(ctxBase)
	defer ccancel()

	// ctx, ccancel := chromedp.NewContext(ctxBase)
	tcontext, tcancel := context.WithTimeout(ccontext, time.Duration(timeout)*time.Second)
	defer tcancel()

	getHTTPResponseHeaders(tcontext, item)

	// Visit the URL
	err := chromedp.Run(tcontext,
		chromedp.ActionFunc(func(ctx context.Context) error {
			if cookie != "" {
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

				var cookieName, cookieValue string
				cookieName = strings.Split(cookie, "=")[0]
				cookieValue = strings.Split(cookie, "=")[1]
				domain := strings.Split(url, "/")[2]
				// logger.Printf("Cookie info %s %s %s\n", cookieName, cookieValue, domain)

				err := network.SetCookie(cookieName, cookieValue).
					WithExpires(&expr).
					WithDomain(domain).
					WithHTTPOnly(true).
					Do(ctx)
				if err != nil {
					return fmt.Errorf("could not set cookie %q", cookie)
				}
			}
			return nil
		}),
		chromedp.Navigate(url),
	)
	if err != nil {
		logger.Fprintln(logger.Writer(), "[+] Error visiting the url : ", url, " - ", err)
		return
	}

	// buffer
	var outerHTML string

	logger.Fprintln(logger.Writer(), "[+] Taking a the HTML content of ", url)

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(tcontext, takeHTML(url, &outerHTML)); err != nil {
		if strings.HasPrefix(err.Error(), "context deadline exceeded") {
			logger.Fprintf(logger.Writer(), "[!] Timeout error for URL %s - %s\n", url, err)
		} else {
			logger.Fprintln(logger.Writer(), "[!] Error in chromedp. Run for URL ", url, " : ", err)
		}
		logger.Fprintln(logger.Writer(), "[-] Retry on :", url)

		// Retry
		if err := chromedp.Run(tcontext, takeHTML(url, &outerHTML)); err != nil {
			if strings.HasPrefix(err.Error(), "context deadline exceeded") {
				logger.Fprintf(logger.Writer(), "[!] 2nd time, timeout error for URL %s - %s\n", url, err)
			} else {
				logger.Fprintln(logger.Writer(), "[!] 2nd time, error in chromedp.Run for URL ", url, " : ", err)
			}
		}
		return
	}

	// Check if the screenshot was taken
	if len(outerHTML) == 0 {
		logger.Fprintln(logger.Writer(), "[!] Error, HTML content not taken for ", url, " because it had a size of 0 bytes")
		return
	}
	filename := filepath.Join(directory, GetHTMLFileName(url))

	if err := os.WriteFile(filename, []byte(outerHTML), 0644); err != nil {
		logger.Fprintln(logger.Writer(), "Error in os.WriteFile ", err, " for url ", url, " with filename ", filename, " and size of ", len(outerHTML))
		return
	}

	logger.Fprintln(logger.Writer(), "[+] Took the HTML content of ", url, " - ", filename, " with a size of ", len(outerHTML))
}

func takeHTML(_ string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Sleep(2 * time.Second),
		chromedp.OuterHTML(`html`, html),
	}
}

// GetHTMLFileName Compute the filename based on the URL.
func GetHTMLFileName(url string) string {
	return GetFileName(url) + ".html"
}
