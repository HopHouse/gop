package gopdynamiccrawler

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	gopstaticcrawler "github.com/hophouse/gop/gopStaticCrawler"
	"github.com/hophouse/gop/screenshot"
	"github.com/hophouse/gop/utils"

	"github.com/PuerkitoBio/goquery"
)

type URLVisitedStruct struct {
	sync.RWMutex
	slice []string
}

var (
	Internal_ressources []gopstaticcrawler.Ressource
	External_ressources []gopstaticcrawler.Ressource
	URLVisited          URLVisitedStruct
	ScreenshotList      []screenshot.Screenshot
	ScreenshotChan      chan struct{}
	UrlChan             chan string
)

func InitCrawler() {
	// Define initial variables
	Internal_ressources = make([]gopstaticcrawler.Ressource, 0)
	External_ressources = make([]gopstaticcrawler.Ressource, 0)
	URLVisited.slice = make([]string, 0)
	ScreenshotList = make([]screenshot.Screenshot, 0)
	ScreenshotChan = make(chan struct{}, *GoCrawlerOptions.ConcurrencyPtr)
	UrlChan = make(chan string)
}

func workerVisit() {
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
	defer cancelBase()

	ctx, cancel := chromedp.NewContext(ctxBase)
	defer cancel()

	tctx, tcancel := context.WithTimeout(ctx, 30*time.Second)
	defer tcancel()

	urlFailed := make([]string, 0)

	nctx, ccancel := chromedp.NewContext(tctx)
	defer ccancel()

	for urlItem := range UrlChan {

		// Retrieve the HTML
		var html string

		err := chromedp.Run(nctx, visitUrlTask(urlItem, &html)...)
		if err != nil {
			utils.Log.Printf("[-] Error with chrome context for url %s", urlItem)

			// If an error was already spotted for this URL
			alreadyPresent := false
			for _, i := range urlFailed {
				if i == urlItem {
					alreadyPresent = true
					break
				}
			}

			if alreadyPresent {
				utils.Log.Printf("[-] Error with chrome context for url %s for the second time. Giving up with this URL.", urlItem)
				utils.CrawlerBar.Done()
				continue
			}

			// If alreay proceed by an other goroutine
			alreadyTreated := false
			URLVisited.RLock()
			for _, i := range URLVisited.slice {
				if i == urlItem {
					alreadyTreated = true
					URLVisited.RUnlock()
					break
				}
			}
			URLVisited.RUnlock()

			if alreadyTreated {
				utils.Log.Printf("[-] Error with chrome context for url %s. URL already treated by an other goroutine.", urlItem)
				utils.CrawlerBar.Done()
				continue
			}

			// First time an error was get, so the url is submitted again
			urlFailed = append(urlFailed, urlItem)
			go func() {
				UrlChan <- urlItem
			}()

			continue
		}

		URLVisited.Lock()
		URLVisited.slice = append(URLVisited.slice, urlItem)
		URLVisited.Unlock()

		// Send it to the treat url
		htmlReader := strings.NewReader(html)
		doc, err := goquery.NewDocumentFromReader(htmlReader)
		if err != nil {
			utils.Log.Printf("[!] Error parsing goquery document for url %s", urlItem)
			utils.CrawlerBar.Done()
			continue
		}

		doc.Url, _ = url.Parse(urlItem)

		results := make([]string, 0)
		results = append(results, TreatA(goquery.CloneDocument(doc))...)
		results = append(results, TreatLinkHref(goquery.CloneDocument(doc))...)
		results = append(results, TreatScriptSrc(goquery.CloneDocument(doc))...)

		uniqueResultsMap := make(map[string]int)
		uniqueResults := make([]string, 0)

		for _, i := range results {
			if i == "" {
				continue
			}

			if _, exist := uniqueResultsMap[i]; !exist {
				uniqueResultsMap[i] = 1
				uniqueResults = append(uniqueResults, i)
			} else {
				uniqueResultsMap[i]++
			}
		}

		utils.CrawlerBar.Add(len(uniqueResults))

		go func() {
			for _, i := range uniqueResults {
				UrlChan <- i
			}
		}()

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

func TreatA(doc *goquery.Document) []string {
	results := make([]string, 0)

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
			URLVisited.RLock()
			for _, item := range URLVisited.slice {
				if item == link {
					//utils.Log.Printf("[*] Url %s already present", link)
					URLVisited.RUnlock()
					return
				}
			}
			URLVisited.RUnlock()

			// Check if the domain is the good one
			linkUrl, _ := url.Parse(link)
			if doc.Url.Host != linkUrl.Host {
				//utils.Log.Printf("[*] Url %s is not the same domain", linkUrl.Host)
				return
			}

			if isInternal == true {
				gopstaticcrawler.PrintNewRessourceFound("internal", "link", link)
			} else {
				gopstaticcrawler.PrintNewRessourceFound("external", "link", link)
			}
		}

		if isAdded && isInternal {
			results = append(results, link)
		}
		return
	})
	return results
}

func TreatScriptSrc(doc *goquery.Document) []string {
	results := make([]string, 0)

	url := doc.Url
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		original_item, exist := s.Attr("src")
		if !exist {
			return
		}

		if strings.HasPrefix(original_item, "javascript:void") {
			//utils.Log.Printf("[-] Not using this script %s from %s\n", original_item, url)
			return
		}

		item := getAbsoluteURL(original_item, url.String())

		result := treatRessource(item, url)
		if item != "" {
			results = append(results, result)
		}
	})

	return results
}

func TreatLinkHref(doc *goquery.Document) []string {
	results := make([]string, 0)
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		original_item, ok := s.Attr("href")
		if !ok {
			return
		}
		url := doc.Url

		item := getAbsoluteURL(original_item, url.String())

		result := treatRessource(item, url)
		if item != "" {
			results = append(results, result)
		}
	})

	return results
}

// Tranform relative path to absolute path if needed and return url
func getAbsoluteURL(original_item string, urlItem string) string {
	item := original_item

	domain := strings.Join(strings.Split(urlItem, "/")[:3], "/")

	if strings.HasPrefix(item, "../") {
		item = domain + "/" + item
		utils.Log.Printf("[*] Transformed from %s to %s\n", original_item, item)
	}

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
func treatRessource(item string, url *url.URL) string {
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
		URLVisited.RLock()
		for _, i := range URLVisited.slice {
			if i == item {
				//utils.Log.Printf("Url %s already present", item)
				URLVisited.RUnlock()
				return ""
			}
		}
		URLVisited.RUnlock()

		// Check if the domain is the good one
		itemUrl, _ := url.Parse(item)
		if url.Host == itemUrl.Host {
			//utils.Log.Printf("Url %s is not the same domain", item)
			return ""
		}

		if isInternal == true {
			gopstaticcrawler.PrintNewRessourceFound("internal", scriptKind, item)
		} else {
			gopstaticcrawler.PrintNewRessourceFound("external", scriptKind, item)
		}
	}

	if isAdded && isInternal {
		return item
	}

	return ""
}
