package gopstaticcrawler

import (
	"os"

	"github.com/gookit/color"
	"github.com/hophouse/gop/utils"
)

var (
	GoCrawlerOptions Options
)

type Options struct {
	UrlPtr         *string
	LogFile        *os.File
	ReportPtr      *bool
	RecursivePtr   *bool
	ScreenshotPtr  *bool
	CookiePtr      *string
	ProxyPtr       *string
	DelayPtr       *int
	ConcurrencyPtr *int
}

func NewOptions(url *string, logFileOption *os.File, report *bool, recursive *bool, screenshot *bool, cookie *string, proxy *string, delay *int, concurrency *int) {
	GoCrawlerOptions.UrlPtr = url
	GoCrawlerOptions.LogFile = logFileOption
	GoCrawlerOptions.ReportPtr = report
	GoCrawlerOptions.RecursivePtr = recursive
	GoCrawlerOptions.ScreenshotPtr = screenshot
	GoCrawlerOptions.CookiePtr = cookie
	GoCrawlerOptions.ProxyPtr = proxy
	GoCrawlerOptions.DelayPtr = delay
	GoCrawlerOptions.ConcurrencyPtr = concurrency
}

func PrintOptions(options *Options) {
	// Custom log file option
	if options.LogFile != nil {
		utils.Log.Println("Using the following log file : ", options.LogFile.Name())
		color.Printf("[+] Using the following log file : %s\n", options.LogFile.Name())
	}

	// Proxy option
	if *options.ProxyPtr != "" {
		utils.Log.Println("Using the following proxy URL", *options.ProxyPtr)
		color.Printf("[+] Using the following proxy URL %s\n", *options.ProxyPtr)
	}

	// Cookie option
	if *options.CookiePtr != "" {
		utils.Log.Println("Using the following custom cookie ", *options.CookiePtr)
		color.Printf("[+] Using the following custom cookie %s\n", *options.CookiePtr)
	}

	// Report option
	if *options.ReportPtr == true {
		utils.Log.Println("Report option set")
		color.Println("[+] Report option set")
	}

	// Screenshot option
	if *options.ScreenshotPtr != false {
		utils.Log.Println("Using the screenshot option")
	}

	// Delay option
	if *options.DelayPtr != 0 {
		utils.Log.Printf("Delay option between requests sets to: %d\n", *options.DelayPtr)
		color.Printf("[+] Delay option between requests sets to: %d\n", *options.DelayPtr)
	}
}
