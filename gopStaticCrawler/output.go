package gopstaticcrawler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/hophouse/gop/utils/logger"
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

	logger.Printf(color.Sprintf("\n"))
	for i := 0; i < len(bannerCrawlerSplited); i++ {
		logger.Printf(color.Sprintf("%s%s\n", Yellow(bannerGoSplited[i]), Cyan(bannerCrawlerSplited[i])))
	}
	logger.Printf(color.Sprintf("\n"))
}

func PrintNewRessourceFound(isInternal, ressourceType, link string) {
	fmt.Fprintf(logger.Writer(), "[+] [%s] [%s] %s\n", isInternal, ressourceType, link)
}

func PrintRessourcesResume(ressourceType string, url string, ressources []*Ressource) {
	logger.Printf(color.Sprintf("\n %s ressources for %s\n", ressourceType, Yellow(url)))
	PrintRessourceList(ressources)
}

func PrintStatistics(duration time.Duration, internal_ressources []*Ressource, external_ressources []*Ressource) {
	var counterStyle, counterScript, counterLink, counterImage, counterUnknown int

	counterStyle, counterScript, counterLink, counterImage, counterUnknown = 0, 0, 0, 0, 0
	logger.Printf(color.Sprintf("\n[+] Statistics\n"))
	logger.Printf(color.Sprintf(" -  Number of internal resources: %s\n", Cyan(len(internal_ressources))))
	for _, item := range internal_ressources {
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
	logger.Printf(color.Sprintf("    - Number of links:    %s\n", Cyan(counterLink)))
	logger.Printf(color.Sprintf("    - Number of scripts:  %s\n", Cyan(counterScript)))
	logger.Printf(color.Sprintf("    - Number of styles:   %s\n", Cyan(counterStyle)))
	logger.Printf(color.Sprintf("    - Number of images:   %s\n", Cyan(counterImage)))
	logger.Printf(color.Sprintf("    - Number of unknowns: %s\n", Cyan(counterUnknown)))

	counterStyle, counterScript, counterLink, counterImage, counterUnknown = 0, 0, 0, 0, 0
	logger.Printf(color.Sprintf("\n -  Number of external resources: %s\n", Cyan(len(external_ressources))))
	for _, item := range external_ressources {
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
	logger.Printf(color.Sprintf("    - Number of links:    %s\n", Cyan(counterLink)))
	logger.Printf(color.Sprintf("    - Number of scripts:  %s\n", Cyan(counterScript)))
	logger.Printf(color.Sprintf("    - Number of styles:   %s\n", Cyan(counterStyle)))
	logger.Printf(color.Sprintf("    - Number of images:   %s\n", Cyan(counterImage)))
	logger.Printf(color.Sprintf("    - Number of unknowns: %s\n", Cyan(counterUnknown)))

	logger.Printf(color.Sprintf("\n -  Execution time: %s\n", Cyan(duration)))
}
