package gopchromedp

import (
	"encoding/json"
	"fmt"

	"github.com/gobuffalo/packr/v2"
	"github.com/hophouse/gop/utils"
)

// GetScreenshotHTML Create the HTML page that references all the taken screenshots.
func ExportHTMLPage() string {
	var htmlCode string

	box := packr.New("Application", "./template")

	htmlCode, err := box.FindString("base.html")
	if err != nil {
		utils.Log.Fatal(err)
	}

	return htmlCode
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
		utils.Log.Println("Error :", err)
	}
	return string(b)
}
