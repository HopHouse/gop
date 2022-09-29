package gopdynamiccrawler

import (
	"os"

	"github.com/gookit/color"
	"github.com/hophouse/gop/utils/logger"
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
		logger.Printf(color.Sprintf("[+] Using the following log file : %s\n", options.LogFile.Name()))
	}

	// Proxy option
	if *options.ProxyPtr != "" {
		logger.Printf(color.Sprintf("[+] Using the following proxy URL %s\n", *options.ProxyPtr))
	}

	// Cookie option
	if *options.CookiePtr != "" {
		logger.Printf(color.Sprintf("[+] Using the following custom cookie %s\n", *options.CookiePtr))
	}

	// Report option
	if *options.ReportPtr == true {
		logger.Println("[+] Report option set")
	}

	// Screenshot option
	if *options.ScreenshotPtr != false {
		logger.Println("Using the screenshot option")
	}

	// Delay option
	if *options.DelayPtr != 0 {
		logger.Printf(color.Sprintf("[+] Delay option between requests sets to: %d\n", *options.DelayPtr))
	}
}
