package gopactivecrawler

import (
	"strings"
	"time"

    "github.com/hophouse/gop/utils"
	"github.com/gookit/color"
)

var (
    Red = color.FgRed.Render
    Green = color.FgGreen.Render
    Yellow = color.FgYellow.Render
    Blue = color.FgBlue.Render
    Cyan = color.FgCyan.Render
    Magenta = color.FgMagenta.Render
)

func PrintOptions(options *Options) () {
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

func PrintBanner() () {
    bannerGo := `
    _________
    __  ____/______
    _  / __  _  __ \
    / /_/ /  / /_/ /
    \____/   \____/
    `
    bannerCrawler := `
          _________                           ______
    __  ____/______________ ____      _____  /_____ ________
     _  /     __  ___/_  __ /__ | /| / /__  / _  _ \__  ___/
    / /___   _  /    / /_/ / __ |/ |/ / _  /  /  __/_  /
    \____/   /_/     \__,_/  ____/|__/  /_/   \___/ /_/
    `
    bannerGoSplited := strings.Split(bannerGo, "\n")
    bannerCrawlerSplited := strings.Split(bannerCrawler, "\n")

    color.Printf("\n")
    for i := 0; i < len(bannerCrawlerSplited) ; i++ {
        color.Printf("%s%s\n", Yellow(bannerGoSplited[i]), Cyan(bannerCrawlerSplited[i]))
    }
    color.Printf("\n")
}

func PrintNewRessourceFound(ressourceType string, link string) () {
	utils.Log.Printf("[+] Found a %s - %s\n", ressourceType, link)
}

func PrintRessourcesResume(ressourceType string, url string, ressources *[]Ressource) () {
    color.Printf("\n %s ressources for %s\n", ressourceType, Yellow(url))
    PrintRessourceList(*ressources)
}

func PrintStatistics(duration time.Duration, internal_ressources *[]Ressource, external_ressources *[]Ressource) () {
    var counterStyle, counterScript, counterLink, counterUnknown int

    counterStyle, counterScript, counterLink, counterUnknown = 0, 0, 0, 0
    color.Printf("\n[+] Statistics\n")
    color.Printf(" -  Number of internal ressources: %s\n", Cyan(len(*internal_ressources)))
    for _, item := range *internal_ressources {
        switch item.Type {
            case "link":
                counterLink += 1
            case "script":
                counterScript += 1
            case "style":
                counterStyle += 1
            default:
                counterUnknown += 1
        }
    }
    color.Printf("    - Number of link:    %s\n", Cyan(counterLink))
    color.Printf("    - Number of script:  %s\n", Cyan(counterScript))
    color.Printf("    - Number of style:   %s\n", Cyan(counterStyle))
    color.Printf("    - Number of unknown: %s\n", Cyan(counterUnknown))

    counterStyle, counterScript, counterLink, counterUnknown = 0, 0, 0, 0
    color.Printf("\n -  Number of external ressources: %s\n", Cyan(len(*external_ressources)))
    for _, item := range *external_ressources {
        switch item.Type {
            case "link":
                counterLink += 1
            case "script":
                counterScript += 1
            case "style":
                counterStyle += 1
            default:
                counterUnknown += 1
        }
    }
    color.Printf("    - Number of link:    %s\n", Cyan(counterLink))
    color.Printf("    - Number of script:  %s\n", Cyan(counterScript))
    color.Printf("    - Number of style:   %s\n", Cyan(counterStyle))
    color.Printf("    - Number of unknown: %s\n", Cyan(counterUnknown))

    color.Printf("\n -  Execution time: %s\n", Cyan(duration))
}