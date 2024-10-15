package gopchromedp

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/hophouse/gop/utils/logger"
)

var (
	//go:embed template/*
	box embed.FS
)

// GetScreenshotHTML Create the HTML page that references all the taken screenshots.
func ExportHTMLPage() string {
	htmlCode, err := box.ReadFile("template/base.html")
	if err != nil {
		logger.Fatal(err)
	}

	htmlCodeStr := string(htmlCode[:])

	return htmlCodeStr
}

func ExportLoadedResources(items []Item) string {
	text := ""

	for _, item := range items {
		for _, resource := range item.Resources {
			text += fmt.Sprintf("[%s] %s\n", item.Url, resource.Url)
		}
	}
	return text
}

func ExportItemsToJSON(items []Item) string {
	b, err := json.Marshal(items)
	if err != nil {
		logger.Println("Error :", err)
	}
	return string(b)
}
