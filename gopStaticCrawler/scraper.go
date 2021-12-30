package gopstaticcrawler

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hophouse/gop/gopchromedp"
	"github.com/hophouse/gop/utils"

	"github.com/gocolly/colly"
)

var (
	Internal_ressources []Ressource
	External_ressources []Ressource
	URLVisited          []string
	ScreenshotList      []gopchromedp.Item
	ConcurrencyChan     chan struct{}
)

func InitCrawler() *colly.Collector {
	c := colly.NewCollector()

	defineCallBacks(c)

	Internal_ressources = make([]Ressource, 0)
	External_ressources = make([]Ressource, 0)
	URLVisited = make([]string, 0)
	ScreenshotList = make([]gopchromedp.Item, 0)
	ConcurrencyChan = make(chan struct{}, *GoCrawlerOptions.ConcurrencyPtr)

	t := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Trust all certificates
	c.WithTransport(&t)
	t.DisableKeepAlives = true

	// Set user-agent
	userAgent := "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0"
	c.UserAgent = userAgent

	if *GoCrawlerOptions.ProxyPtr != "" {
		url, err := url.Parse(*GoCrawlerOptions.ProxyPtr)
		if err != nil {
			utils.Log.Fatalf("Error with proxy: %s", err)
		}
		c.SetProxy(url.String())
		c.SetProxyFunc(http.ProxyURL(url))
	}

	if *GoCrawlerOptions.DelayPtr != 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob: "*",
			Delay:      time.Duration(*GoCrawlerOptions.DelayPtr) * time.Second,
		})
	}
	return c
}

func VisiteURL(visited *[]string, c *colly.Collector, Url string) {
	cleanedUrl := Url

	// Remove GET parameters
	if strings.Contains(Url, "?") == true {
		cleanedUrl = strings.Split(Url, "?")[0]
	}

	// Check if the page was already visited
	for _, item := range *visited {
		if item == cleanedUrl || item == Url {
			return
		}
	}

	// Check if we are in a loop
	s := strings.Split(cleanedUrl, "/")
	length := len(s)
	if len(s) > 3 && len(unique(s)) < (length-2) {
		utils.Log.Println("[-] This page might be a redirection ", Url, ", so we do not visite it.")
		return
	}

	// Check if the page will logout and potentialy remove the token passed in parameter
	if *GoCrawlerOptions.CookiePtr != "" {
		if strings.Contains(Url, "logout") || strings.Contains(Url, "deconnexion") {
			utils.Log.Printf("[-] This URL %s might contains a logout URL that may invalidate the session cookie. It will not be proceeded.", Url)
			return
		}

	}
	c.Visit(cleanedUrl)

	// Add to visited URL
	*visited = append(*visited, cleanedUrl)
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

func defineCallBacks(c *colly.Collector) {
	c.OnRequest(func(r *colly.Request) {
		utils.Log.Printf("[+] Sending request to %s\n", r.URL)
		utils.CrawlerBar.AddAndIncrementTotal(1)
		if *GoCrawlerOptions.CookiePtr != "" {
			r.Headers.Set("cookie", *GoCrawlerOptions.CookiePtr)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		utils.Log.Printf("[/] Response from %s: %d\n", r.Request.URL, r.StatusCode)

		// Take a screenshot if the option was set
		if *GoCrawlerOptions.ScreenshotPtr == true {
			utils.ScreenshotBar.AddAndIncrementTotal(1)

			go func() {
				item := gopchromedp.NewItem(r.Request.URL.String())
				gopchromedp.TakeScreenShot(&item, "screenshots/", *GoCrawlerOptions.ProxyPtr, *GoCrawlerOptions.CookiePtr, *GoCrawlerOptions.DelayPtr)

				// Add screenshot to list
				ScreenshotList = append(ScreenshotList, item)

				<-ConcurrencyChan
			}()
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		original_link := e.Attr("href")
		link := e.Attr("href")
		url := e.Request.URL

		if strings.HasPrefix(link, "#") || strings.HasPrefix(link, "?") || link == "/" || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") || link == "" {
			utils.Log.Printf("[-] Not using this link %s from %s\n", link, url)
			return
		}

		getAbsoluteURL(&link, original_link, url)

		isInternal, ressource := CreateRessource(url.String(), link, "link")
		if isInternal == true {
			if isAdded := AddRessourceIfDoNotExists(&Internal_ressources, ressource); isAdded {
				PrintNewRessourceFound("internal", "link", link)
			}
		} else {
			if isAdded := AddRessourceIfDoNotExists(&External_ressources, ressource); isAdded {
				PrintNewRessourceFound("external", "link", link)
			}
		}
	})

	c.OnHTML("script[src]", TreatScriptSrc)

	c.OnHTML("link[href]", TreatLinkHref)

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		utils.Log.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "\nError:", err)
		defer utils.CrawlerBar.Done()
	})

	c.OnScraped(func(r *colly.Response) {
		utils.Log.Printf("[+] Finished sending ressources to %s\n", r.Request.URL)
		defer utils.CrawlerBar.Done()
	})
}

func TreatScriptSrc(e *colly.HTMLElement) {
	original_item := e.Attr("src")
	item := e.Attr("src")
	url := e.Request.URL

	if strings.HasPrefix(item, "javascript:void") {
		utils.Log.Printf("[-] Not using this script %s from %s\n", item, url)
		return
	}

	getAbsoluteURL(&item, original_item, url)
	treatRessource(item, url)
}

func TreatLinkHref(e *colly.HTMLElement) {
	original_item := e.Attr("href")
	item := e.Attr("href")
	url := e.Request.URL

	getAbsoluteURL(&item, original_item, url)

	treatRessource(item, url)
}

// Tranform relative path to absolute path if needed and return url
func getAbsoluteURL(item *string, original_item string, url *url.URL) {
	var domain string = strings.Join(strings.Split(url.String(), "/")[:3], "/")
	if strings.HasPrefix(*item, "/") {
		if strings.HasSuffix(url.String(), "/") {
			*item = domain + (*item)[1:]
		} else {
			*item = domain + *item
		}
		utils.Log.Printf("[*] Modified item for URL for %s. Transformed from %s to %s\n", url, original_item, *item)
	}

	if strings.HasPrefix(url.String(), " ") {
		*item = (*item)[1:]
		utils.Log.Printf("[*] Modified %s. Removed space from %s on %s\n", *item, original_item, url)
	}

}

// Treat url, classify what the ressource is and add to the internal or
// external scope
func treatRessource(item string, url *url.URL) {
	var scriptKind = "unknown"

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

	isInternal, ressource := CreateRessource(url.String(), item, scriptKind)
	if isInternal == true {
		if isAdded := AddRessourceIfDoNotExists(&Internal_ressources, ressource); isAdded {
			PrintNewRessourceFound("internal", scriptKind, item)
		}
	} else {
		if isAdded := AddRessourceIfDoNotExists(&External_ressources, ressource); isAdded {
			PrintNewRessourceFound("external", scriptKind, item)
		}
	}
}
