package gopactivecrawler

import (
    "net/http"
    "strings"
    "crypto/tls"
    "net/url"
    "time"

    "github.com/hophouse/gop/utils"
    "github.com/hophouse/gop/screenshot"

    "github.com/gocolly/colly"
)

var (
    Internal_ressources []Ressource
    External_ressources []Ressource
    URLVisited []string
    ScreenshotList []screenshot.Screenshot
    ConcurrencyChan chan struct{}
)

func InitCrawler() (*colly.Collector) {
    c := colly.NewCollector()
    defineCallBacks(c)

    Internal_ressources = make([]Ressource, 0)
    External_ressources = make([]Ressource, 0)
    URLVisited = make([]string, 0)
    ScreenshotList = make([]screenshot.Screenshot, 0)
    ConcurrencyChan = make(chan struct{}, *GoCrawlerOptions.ConcurrencyPtr)


    t := http.Transport{
        TLSClientConfig:&tls.Config{InsecureSkipVerify: true},
    }

    // Trust all certificates
    c.WithTransport(&t)
	t.DisableKeepAlives = true

    // Set user-agent
    userAgent := "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0"
    c.UserAgent = userAgent

    if (*GoCrawlerOptions.ProxyPtr != "") {
        url, err := url.Parse(*GoCrawlerOptions.ProxyPtr)
        if err != nil {
            utils.Log.Fatalf("Error with proxy: %s", err)
        }
        c.SetProxy(url.String())
        c.SetProxyFunc(http.ProxyURL(url))
    }

    if *GoCrawlerOptions.DelayPtr != 0 {
        c.Limit(&colly.LimitRule{
            DomainGlob:  "*",
            Delay: time.Duration(*GoCrawlerOptions.DelayPtr) * time.Second,
        })
    }
    return c
}

func VisiteURL(visited *[]string, c *colly.Collector, Url string) () {
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
    if (len(s) > 3 && len(unique(s)) < (length-2)) {
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

func defineCallBacks(c *colly.Collector) () {
    c.OnRequest(func(r *colly.Request) {
        utils.Log.Printf("[+] Sending request to %s\n", r.URL)
        utils.CrawlerBar.Add(1)
        if (*GoCrawlerOptions.CookiePtr != "") {
            r.Headers.Set("cookie", *GoCrawlerOptions.CookiePtr)
        }
    })

    c.OnResponse(func(r *colly.Response) {
        utils.Log.Printf("[/] Response from %s: %d\n", r.Request.URL, r.StatusCode)

        // Take a screenshot if the option was set
        if (*GoCrawlerOptions.ScreenshotPtr == true) {
            utils.ScreenshotBar.Add(1)

            go func() {
                ConcurrencyChan <- struct{}{}

                screenshot.TakeScreenShot(r.Request.URL.String(), "screenshots/", *GoCrawlerOptions.CookiePtr, *GoCrawlerOptions.ProxyPtr)

                // Add screenshot to list
                ScreenshotList = append(ScreenshotList, screenshot.Screenshot{
                    Url: r.Request.URL.String(),
                    RequestStatus: "Uknown",
                })

                <- ConcurrencyChan
            } ()
        }
    })

    c.OnHTML("a[href]", func(e *colly.HTMLElement) {
        original_link := e.Attr("href")
        link := e.Attr("href")
        url := e.Request.URL

        if (strings.HasPrefix(link, "#") || strings.HasPrefix(link, "?") || link == "/" || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") || link == "") {
            utils.Log.Printf("[-] Not using this link %s from %s\n", link, url)
            return
        }

        var domain string = strings.Join(strings.Split(url.String(), "/")[:3], "/")
        if (strings.HasPrefix(link, "/")) {
            if (strings.HasSuffix(url.String(), "/")) {
                link = domain + link[1:]
            } else {
                link = domain + link
            }
            utils.Log.Printf("[*] Modified %s. Transformed from %s on %s\n", link, original_link, url)
        }

        if (strings.HasPrefix(url.String(), " ")) {
            link = link[1:]
            utils.Log.Printf("[*] Modified %s. Removed space from %s on %s\n", link, original_link, url)
        }

        var isAdded bool

        isInternal, ressource := CreateRessource(*GoCrawlerOptions.UrlPtr, link, "link")
        if (isInternal == true) {
            isAdded = AddRessourceIfDoNotExists(&Internal_ressources, ressource)
            if (*GoCrawlerOptions.RecursivePtr == true) {
                utils.Log.Printf("[+] Adding for visit %s\n", link)
                VisiteURL(&URLVisited, c, link)
            }
        } else {
            isAdded = AddRessourceIfDoNotExists(&External_ressources, ressource)
        }

        if (isAdded == true) {
            PrintNewRessourceFound("link", link)
        }
    })

    c.OnHTML("script[src]", func(e *colly.HTMLElement) {
        original_script := e.Attr("src")
        script := e.Attr("src")
        url := e.Request.URL

        if (strings.HasPrefix(script, "javascript:void")) {
            utils.Log.Printf("[-] Not using this script %s from %s\n", script, url)
            return
        }

        var domain string = strings.Join(strings.Split(url.String(), "/")[:3], "/")
        if (strings.HasPrefix(script, "/")) {
            if (strings.HasSuffix(url.String(), "/")) {
                script = domain + script[1:]
            } else {
                script = domain + script
            }
            utils.Log.Printf("[*] Modified script for URL for %s. Transformed from %s to %s\n", url, original_script, script)
        }

        var isAdded bool

        isInternal, ressource := CreateRessource(url.String(), script, "script")

        if (isInternal == true) {
            isAdded = AddRessourceIfDoNotExists(&Internal_ressources, ressource)
        } else {
            isAdded = AddRessourceIfDoNotExists(&External_ressources, ressource)
        }

        if (isAdded == true) {
            PrintNewRessourceFound("script", script)
        }
    })

    c.OnHTML("link[href]", func(e *colly.HTMLElement) {
        original_style := e.Attr("href")
        style := e.Attr("href")
        url := e.Request.URL

        var domain string = strings.Join(strings.Split(url.String(), "/")[:3], "/")
        if (strings.HasPrefix(style, "/")) {
            if (strings.HasSuffix(style, "/")) {
                style = domain + style[1:]
            } else {
                style = domain + style
            }
            utils.Log.Printf("[*] Modified script for URL for %s. Transformed from %s to %s\n", url, original_style, style)
        }

        var isAdded bool
        isInternal, ressource := CreateRessource(url.String(), style, "style")
        if (isInternal == true) {
            isAdded = AddRessourceIfDoNotExists(&Internal_ressources, ressource)
        } else {
            isAdded = AddRessourceIfDoNotExists(&External_ressources, ressource)
        }

        if (isAdded){
            PrintNewRessourceFound("style file", style)
        }
    })

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
