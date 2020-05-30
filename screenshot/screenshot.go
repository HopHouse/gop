package screenshot

import (
    "context"
    "io/ioutil"
    "fmt"
    "math"
    "strings"
    "time"

    "github.com/hophouse/gop/utils"
    "github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
    "github.com/chromedp/cdproto/network"
    "github.com/chromedp/cdproto/cdp"
    "github.com/chromedp/chromedp"
)

type Screenshot struct {
    Url string
    RequestStatus string
}

// Take screenshot of the pages
func TakeScreenShot(url string, directory string, cookie string, proxy string) () {
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
        if (strings.HasPrefix(err.Error(), "context deadline exceeded")) {
            utils.Log.Printf("[!] Timeout error for URL %s\n", url)
        } else {
            utils.Log.Println("[!] Error in chromedp. Run for URL ", url, " : ", err)
        }
        utils.Log.Println("[-] Retry on :", url)

        // Retry
        if err := chromedp.Run(tctx, fullScreenshot(url, cookie, 90, &buf)); err != nil {
            if (strings.HasPrefix(err.Error(), "context deadline exceeded")) {
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

    filename := directory + strings.ReplaceAll(strings.ReplaceAll(url, "/", ""), ":", "_") + ".png"
    if err := ioutil.WriteFile(filename, buf, 0644); err != nil {
        utils.Log.Println("Error in ioutil.WriteFile", err, "for url: ", url)
        return
    }
    utils.Log.Println("[+] Took a screenshot of ", url, " - ", filename)
}

func GetScreenshotFileName(url string) string {
    filename := strings.ReplaceAll(url, ":", "_")
    filename = strings.ReplaceAll(filename, "/", "")
    filename += ".png"
    return filename
}

// fullScreenshot takes a screenshot of the entire browser viewport.
// Liberally copied from puppeteer's source.
func fullScreenshot(urlstr string, cookie string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
            // add cookies to chrome
            if  cookie != "" {
                // create cookie expiration
                expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

                var cookieName, cookieValue string
                cookieName = strings.Split(cookie, "=")[0]
                cookieValue = strings.Split(cookie, "=")[1]
                domain := strings.Split(urlstr, "/")[2]

                success, err := network.SetCookie(cookieName, cookieValue).
                    WithExpires(&expr).
                    WithDomain(domain).
                    WithHTTPOnly(true).
                    Do(ctx)
                if err != nil {
                    return err
                }
                if !success {
                    return fmt.Errorf("could not set cookie %q", cookie)
                }
            }
			return nil
        }),
        chromedp.Navigate(urlstr),
        chromedp.Sleep(20 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
                utils.Log.Println("Error in page.GetLayoutMetrics ", err)
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
                Do(ctx)
			if err != nil {
				return err
                utils.Log.Println("Error in emulation.SetDeviceMetricsOverride ", err)
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