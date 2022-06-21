package gopstaticcrawler

import (
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/hophouse/gop/utils"
)

var (
	Red     = color.FgRed.Render
	Green   = color.FgGreen.Render
	Yellow  = color.FgYellow.Render
	Blue    = color.FgBlue.Render
	Cyan    = color.FgCyan.Render
	Magenta = color.FgMagenta.Render
)

func PrintBanner() {
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
	for i := 0; i < len(bannerCrawlerSplited); i++ {
		color.Printf("%s%s\n", Yellow(bannerGoSplited[i]), Cyan(bannerCrawlerSplited[i]))
	}
	color.Printf("\n")
}

func PrintNewRessourceFound(isInternal, ressourceType, link string) {
	utils.Log.Printf("[+] [%s] [%s] %s\n", isInternal, ressourceType, link)
}

func PrintRessourcesResume(ressourceType string, url string, ressources *[]Ressource) {
	color.Printf("\n %s ressources for %s\n", ressourceType, Yellow(url))
	PrintRessourceList(*ressources)
}

func PrintStatistics(duration time.Duration, internal_ressources *[]Ressource, external_ressources *[]Ressource) {
	var counterStyle, counterScript, counterLink, counterImage, counterUnknown int

	counterStyle, counterScript, counterLink, counterImage, counterUnknown = 0, 0, 0, 0, 0
	color.Printf("\n[+] Statistics\n")
	color.Printf(" -  Number of internal resources: %s\n", Cyan(len(*internal_ressources)))
	for _, item := range *internal_ressources {
		switch item.Type {
		case "link":
			counterLink += 1
		case "script":
			counterScript += 1
		case "style":
			counterStyle += 1
		case "image":
			counterImage += 1
		default:
			counterUnknown += 1
		}
	}
	color.Printf("    - Number of links:    %s\n", Cyan(counterLink))
	color.Printf("    - Number of scripts:  %s\n", Cyan(counterScript))
	color.Printf("    - Number of styles:   %s\n", Cyan(counterStyle))
	color.Printf("    - Number of images:   %s\n", Cyan(counterImage))
	color.Printf("    - Number of unknowns: %s\n", Cyan(counterUnknown))

	counterStyle, counterScript, counterLink, counterImage, counterUnknown = 0, 0, 0, 0, 0
	color.Printf("\n -  Number of external resources: %s\n", Cyan(len(*external_ressources)))
	for _, item := range *external_ressources {
		switch item.Type {
		case "link":
			counterLink += 1
		case "script":
			counterScript += 1
		case "style":
			counterStyle += 1
		case "image":
			counterImage += 1
		default:
			counterUnknown += 1
		}
	}
	color.Printf("    - Number of links:    %s\n", Cyan(counterLink))
	color.Printf("    - Number of scripts:  %s\n", Cyan(counterScript))
	color.Printf("    - Number of styles:   %s\n", Cyan(counterStyle))
	color.Printf("    - Number of images:   %s\n", Cyan(counterImage))
	color.Printf("    - Number of unknowns: %s\n", Cyan(counterUnknown))

	color.Printf("\n -  Execution time: %s\n", Cyan(duration))
}
