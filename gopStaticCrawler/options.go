package gopstaticcrawler

import (
	"os"
)

var (
    GoCrawlerOptions Options
)

type Options struct {
    UrlPtr *string
    LogFile *os.File
    ReportPtr *bool
    RecursivePtr *bool
    ScreenshotPtr *bool
    CookiePtr *string
    ProxyPtr *string
    DelayPtr *int
    ConcurrencyPtr *int
}

func NewOptions(url *string, logFileOption *os.File, report *bool, recursive *bool, screenshot *bool, cookie *string, proxy *string, delay *int, concurrency *int) () {
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