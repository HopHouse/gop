package gopdynamiccrawler

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	gopstaticcrawler "github.com/hophouse/gop/gopStaticCrawler"
	"github.com/hophouse/gop/screenshot"
	"github.com/hophouse/gop/utils"

	"github.com/PuerkitoBio/goquery"
)

var (
	Internal_ressources []gopstaticcrawler.Ressource
	External_ressources []gopstaticcrawler.Ressource
	URLVisited          []string
	ScreenshotList      []screenshot.Screenshot
	ScreenshotChan      chan struct{}
	UrlChan             chan string
	HtmlChan            chan *goquery.Document
)

func InitCrawler() {
	// Define initial variables
	Internal_ressources = make([]gopstaticcrawler.Ressource, 0)
	External_ressources = make([]gopstaticcrawler.Ressource, 0)
	URLVisited = make([]string, 0)
	ScreenshotList = make([]screenshot.Screenshot, 0)
	ScreenshotChan = make(chan struct{}, *GoCrawlerOptions.ConcurrencyPtr)
	UrlChan = make(chan string)
	HtmlChan = make(chan *goquery.Document)
}

func workerVisit(urlChan chan string, htmlChan chan *goquery.Document) {
	// Trust all certificates
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"),
	)

	// Set a proxy
	if *GoCrawlerOptions.ProxyPtr != "" {
		proxyUrl, err := url.Parse(*GoCrawlerOptions.ProxyPtr)
		if err != nil {
			utils.Log.Fatalf("Error with proxy: %s", err)
		}
		proxyAllocatorOption := chromedp.ProxyServer(proxyUrl.String())
		options = append(options, proxyAllocatorOption)
	}

	ctxBase, cancelBase := chromedp.NewExecAllocator(context.Background(), options...)

	ctx, cancel := chromedp.NewContext(ctxBase)

	tctx, tcancel := context.WithTimeout(ctx, 30*time.Second)


	for urlItem := range urlChan {
		ctx, ccancel := chromedp.NewContext(tctx)
		defer ccancel()

		URLVisited = append(URLVisited, urlItem)

		// Retrieve the HTML
		var html string
		err := chromedp.Run(ctx, visitUrlTask(urlItem, &html)...)
		if err != nil {
			utils.Log.Printf("Error with chrome context for url %s", urlItem)
			utils.CrawlerBar.Done()
			continue
		}

		// Send it to the treat url
		htmlReader := strings.NewReader(html)
		doc, err := goquery.NewDocumentFromReader(htmlReader)
		if err != nil {
			utils.Log.Printf("Error parsing goquery document for url %s", urlItem)
			utils.CrawlerBar.Done()
			continue
		}

		doc.Url, _ = url.Parse(urlItem)

		utils.CrawlerBar.Add(1)
		htmlChan <- doc
		utils.CrawlerBar.Done()
	}

	defer cancelBase()
	defer cancel()
	defer tcancel()
}

func workerParse(htmlChan chan *goquery.Document, urlChan chan string) {
	for doc := range htmlChan {
		TreatA(goquery.CloneDocument(doc))
		TreatLinkHref(goquery.CloneDocument(doc))
		TreatScriptSrc(goquery.CloneDocument(doc))

		utils.CrawlerBar.Done()
	}
}

func visitUrlTask(url string, html *string) []chromedp.Action {
	*html = ""
	
	actions := make([]chromedp.Action, 0)

	actions = append(actions, chromedp.Navigate(url))

	if !strings.HasSuffix(url, ".pdf") {
		actions = append(actions, chromedp.OuterHTML("html", html))
	}

	if *GoCrawlerOptions.DelayPtr != 0 {
		delay := time.Duration(*GoCrawlerOptions.DelayPtr) * time.Second
		actions = append(actions, chromedp.Sleep(delay))
	}

	return actions
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func TreatA(doc *goquery.Document) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		original_link, ok := s.Attr("href")
		if !ok {
			return
		}

		if strings.HasPrefix(original_link, "#") || strings.HasPrefix(original_link, "?") || original_link == "/" || strings.HasPrefix(original_link, "javascript:") || strings.HasPrefix(original_link, "mailto:") || original_link == "" {
			return
		}

		link := getAbsoluteURL(original_link, doc.Url.String())

		var isAdded bool

		isInternal, ressource := gopstaticcrawler.CreateRessource(doc.Url.String(), link, "link")
		if isInternal == true {
			isAdded = gopstaticcrawler.AddRessourceIfDoNotExists(&Internal_ressources, ressource)
		} else {
			isAdded = gopstaticcrawler.AddRessourceIfDoNotExists(&External_ressources, ressource)
		}

		if isAdded {
			// Check if the page was already visited
			for _, item := range URLVisited {
				if item == link {
					return
				}
			}

			// Check if the domain is the good one
			linkUrl, _ := url.Parse(link)
			if doc.Url.Host != linkUrl.Host {
				return
			}

			gopstaticcrawler.PrintNewRessourceFound("link", link)

			UrlChan <- link
			utils.CrawlerBar.Add(1)
		}
	})
}

func TreatScriptSrc(doc *goquery.Document) {
	url := doc.Url
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		original_item, exist := s.Attr("src")
		if !exist {
			return
		}

		if strings.HasPrefix(original_item, "javascript:void") {
			utils.Log.Printf("[-] Not using this script %s from %s\n", original_item, url)
			return
		}

		item := getAbsoluteURL(original_item, url.String())
		treatRessource(item, url)
	})
}

func TreatLinkHref(doc *goquery.Document) {
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		original_item, ok := s.Attr("href")
		if !ok {
			return
		}
		url := doc.Url

		item := getAbsoluteURL(original_item, url.String())

		treatRessource(item, url)
	})
}

// Tranform relative path to absolute path if needed and return url
func getAbsoluteURL(original_item string, urlItem string) string {
	item := original_item

	domain := strings.Join(strings.Split(urlItem, "/")[:3], "/")

	if strings.HasPrefix(item, "/") {
		if strings.HasPrefix(item, "//") {
			item = "https:" + item
			utils.Log.Printf("[*] Transformed from %s to %s\n", original_item, item)
		} else {
			if strings.HasSuffix(urlItem, "/") {
				item = domain + (item)[1:]
			} else {
				item = domain + item
			}
			utils.Log.Printf("[*] Transformed from %s to %s\n", original_item, item)
		}
	}

	if strings.HasPrefix(urlItem, " ") {
		item = (item)[1:]
		utils.Log.Printf("[*] Transformed from %s to %s\n", original_item, item)
	}

	return item
}

// Treat url, classify what the ressource is and add urlto the internal or
// external scope
func treatRessource(item string, url *url.URL) {
	var isAdded bool
	var scriptKind = "link"

	file := strings.Split(item, "?")[0]

	if strings.HasSuffix(file, ".js") || strings.HasSuffix(file, ".jsf") {
		scriptKind = "script"
	}

	if strings.HasSuffix(file, ".png") || strings.HasSuffix(file, ".jpg") || strings.HasSuffix(file, ".ico") {
		scriptKind = "image"
	}

	if strings.HasSuffix(file, ".css") {
		scriptKind = "style"
	}

	isInternal, ressource := gopstaticcrawler.CreateRessource(url.String(), item, scriptKind)
	if isInternal == true {
		isAdded = gopstaticcrawler.AddRessourceIfDoNotExists(&Internal_ressources, ressource)
	} else {
		isAdded = gopstaticcrawler.AddRessourceIfDoNotExists(&External_ressources, ressource)
	}

	if isAdded {
		// Check if the page was already visited
		for _, i := range URLVisited {
			if i == item {
				return
			}
		}

		// Check if the domain is the good one
		itemUrl, _ := url.Parse(item)
		if url.Host == itemUrl.Host {
			return
		}

		gopstaticcrawler.PrintNewRessourceFound(scriptKind, item)
		UrlChan <- item
		utils.CrawlerBar.Add(1)
	}
}
