package screenshot

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gobuffalo/packr"
	"github.com/hophouse/gop/utils"
)

// Screenshot structure to defile the Url and the status when requested
type Screenshot struct {
	Url           string
	RequestStatus string
}

// TakeScreenShot Take screenshot of the pages.
func TakeScreenShot(url string, directory string, cookie string, proxy string) {
	defer utils.ScreenshotBar.Done()

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"),
	)

	// Init chrome context
	if proxy != "" {
		proxyAllocatorOption := chromedp.ProxyServer(proxy)
		options = append(options, proxyAllocatorOption)
	}

	ctxBase, cancelBase := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancelBase()

	ctx, cancel := chromedp.NewContext(ctxBase)
	defer cancel()

	tctx, tcancel := context.WithTimeout(ctx, 1*time.Minute)
	defer tcancel()

	// buffer
	var buf []byte

	// take screenshot for all urls
	if strings.HasSuffix(url, ".pdf") {
		utils.Log.Println("[+] Do not take a screenshot of the PDF ", url)
		return
	}
	// utils.Log.Println("[+] Taking a screenshot of ", url)

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(tctx, fullScreenshot(url, cookie, 90, &buf)); err != nil {
		if strings.HasPrefix(err.Error(), "context deadline exceeded") {
			utils.Log.Printf("[!] Timeout error for URL %s\n", url)
		} else {
			utils.Log.Println("[!] Error in chromedp. Run for URL ", url, " : ", err)
		}
		utils.Log.Println("[-] Retry on :", url)

		// Retry
		if err := chromedp.Run(tctx, fullScreenshot(url, cookie, 90, &buf)); err != nil {
			if strings.HasPrefix(err.Error(), "context deadline exceeded") {
				utils.Log.Printf("[!] 2nd time, timeout error for URL %s\n", url)
			} else {
				utils.Log.Println("[!] 2nd time, error in chromedp.Run for URL ", url, " : ", err)
			}
		}
		return
	}

	// Check if the screenshot was taken
	if len(buf) == 0 || len(buf) == 3249 {
		utils.Log.Println("[!] Error, screenshot not taken for ", url, " because it had a size of 0 bytes")
		return
	}
	filename := filepath.Join(directory, GetScreenshotFileName(url))

	if err := ioutil.WriteFile(filename, buf, 0644); err != nil {
		utils.Log.Println("Error in ioutil.WriteFile ", err, " for url ", url, " with filename ", filename, " and size of ", len(buf))
		return
	}
	utils.Log.Println("[+] Took a screenshot of ", url, " - ", filename, " with a size of ", len(buf))
}

// GetScreenshotFileName Compute the filename based on the URL.
func GetScreenshotFileName(url string) string {
	filename := strings.ReplaceAll(url, ":", "-")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, ".", "_")
	filename = strings.ReplaceAll(filename, "?", "-")
	filename += ".png"
	return filename
}

// fullScreenshot takes a screenshot of the entire browser viewport.
// Liberally copied from puppeteer's source.
func fullScreenshot(urlstr string, cookie string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// add cookies to chrome
			/*
				if cookie != "" {
					// create cookie expiration
					expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

					var cookieName, cookieValue string
					cookieName = strings.Split(cookie, "=")[0]
					cookieValue = strings.Split(cookie, "=")[1]
					domain := strings.Split(urlstr, "/")[2]
					//fmt.Printf("Cookie info %s %s %s\n", cookieName, cookieValue, domain)

					_, err := network.SetCookie(cookieName, cookieValue).
						WithExpires(&expr).
						WithDomain(domain).
						WithHTTPOnly(true).
						Do(ctx)

					if err != nil {
						return fmt.Errorf("could not set cookie %q", cookie)
					}
				}
			*/
			return nil
		}),
		chromedp.Navigate(urlstr),
		chromedp.Sleep(20 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, _, _, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				utils.Log.Println("Error in page.GetLayoutMetrics ", err)
				return err
			}

			// If content size is empty
			if contentSize == nil {
				contentSize = &dom.Rect{
					X:      0,
					Y:      0,
					Width:  1920,
					Height: 1080,
				}
			}

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(int64(contentSize.Width), int64(contentSize.Height), 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				utils.Log.Println("Error in emulation.SetDeviceMetricsOverride ", err)
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}

			return nil
		}),
	}
}

// GetScreenshotHTML Create the HTML page that references all the taken screenshots.
func GetScreenshotHTML(sl []Screenshot) string {
	var htmlCode string

	box := packr.NewBox("./template")

	htmlHeader, err := box.FindString("base_header.html")
	if err != nil {
		utils.Log.Fatal(err)
	}

	htmlCode += htmlHeader

	for _, item := range sl {
		filename := "./screenshots/" + GetScreenshotFileName(item.Url)

		utils.Log.Println(filename)
		htmlCode += fmt.Sprintf("<div class=\"col-md-4\">\n")
		htmlCode += fmt.Sprintf("  <div class=\"card mb-4 shadow-sm\">\n")

		// Screenshot
		htmlCode += fmt.Sprintf("    <img class=\"bd-placeholder-img card-img-top\" width=\"100%%\" height=\"225\" focusable=\"false\" src=\"%s\" />\n", filename)
		htmlCode += fmt.Sprintf("    <div class=\"card-body\">\n")

		// Request
		htmlCode += fmt.Sprintf("      <p class=\"card-text\">%s</p>\n", item.Url)
		htmlCode += fmt.Sprintf("      <div class=\"d-flex justify-content-between align-items-center\">\n")
		htmlCode += fmt.Sprintf("        <div class=\"btn-group\">\n")
		// Visit
		htmlCode += fmt.Sprintf("          <a href=\"%s\"><button type=\"button\" class=\"btn btn-sm btn-outline-secondary\">Visit</button></a>", item.Url)
		htmlCode += fmt.Sprintf("        </div>\n")
		htmlCode += fmt.Sprintf("      </div>\n")
		htmlCode += fmt.Sprintf("    </div>\n")
		htmlCode += fmt.Sprintf("  </div>\n")
		htmlCode += fmt.Sprintf("</div>\n")
		htmlCode += fmt.Sprintf("\n\n")
	}

	htmlFooter, err := box.FindString("base_footer.html")
	if err != nil {
		utils.Log.Fatal(err)
	}

	htmlCode += htmlFooter

	return htmlCode
}
