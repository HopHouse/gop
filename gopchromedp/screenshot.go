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
	"github.com/hophouse/gop/utils"
)

// TakeScreenShot Take screenshot of the pages.
func TakeScreenShot(item *Item, directory string, proxy string, cookie string, timeout int) {
	// take screenshot for all item.Urls
	if strings.HasSuffix(item.Url, ".pdf") {
		utils.Log.Println("[+] Do not take a screenshot of the PDF ", item.Url)
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

	// Visit the item.Url
	err := chromedp.Run(tcontext,
		chromedp.ActionFunc(func(ctx context.Context) error {
			if cookie != "" {
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

				var cookieName, cookieValue string
				cookieName = strings.Split(cookie, "=")[0]
				cookieValue = strings.Split(cookie, "=")[1]
				domain := strings.Split(item.Url, "/")[2]
				// fmt.Printf("Cookie info %s %s %s\n", cookieName, cookieValue, domain)

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
		chromedp.Navigate(item.Url),
	)
	if err != nil {
		utils.Log.Println("[+] Error visiting the item.Url : ", item.Url, " - ", err)
		return
	}

	// buffer
	var buf []byte

	utils.Log.Println("[+] Taking a screenshot of ", item.Url)

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(tcontext, fullScreenshot(90, &buf)); err != nil {
		if strings.HasPrefix(err.Error(), "context deadline exceeded") {
			utils.Log.Printf("[!] Timeout error for item.Url %s - %s\n", item.Url, err)
		} else {
			utils.Log.Println("[!] Error in chromedp. Run for item.Url ", item.Url, " : ", err)
		}
		utils.Log.Println("[-] Retry on :", item.Url)

		// Retry
		if err := chromedp.Run(tcontext, fullScreenshot(90, &buf)); err != nil {
			if strings.HasPrefix(err.Error(), "context deadline exceeded") {
				utils.Log.Printf("[!] 2nd time, timeout error for item.Url %s - %s\n", item.Url, err)
			} else {
				utils.Log.Println("[!] 2nd time, error in chromedp.Run for item.Url ", item.Url, " : ", err)
			}
		}
		return
	}

	// Check if the screenshot was taken
	if len(buf) == 0 || len(buf) == 3249 {
		utils.Log.Println("[!] Error, screenshot not taken for ", item.Url, " because it had a size of 0 bytes")
		return
	}
	filename := filepath.Join(directory, GetScreenshotFileName(item.Url))

	if err := os.WriteFile(filename, buf, 0644); err != nil {
		utils.Log.Println("Error in os.WriteFile ", err, " for item.Url ", item.Url, " with filename ", filename, " and size of ", len(buf))
		return
	}

	utils.Log.Println("[+] Took a screenshot of ", item.Url, " - ", filename, " with a size of ", len(buf))
}

// GetScreenshotFileName Compute the filename based on the item.Url.
func GetScreenshotFileName(url string) string {
	return GetFileName(url) + ".png"
}

// fullScreenshot takes a screenshot of the entire browser viewport.
func fullScreenshot(quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.FullScreenshot(res, int(quality)),
		/*
			// Liberally copied from puppeteer's source.
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
		*/
	}
}
