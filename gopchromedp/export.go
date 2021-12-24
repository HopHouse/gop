package gopchromedp

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/hophouse/gop/utils"
)

// GetScreenshotHTML Create the HTML page that references all the taken screenshots.
func ExportScreenshotsToHTML(items []Item) string {
	var htmlCode string

	box := packr.NewBox("./template")

	htmlHeader, err := box.FindString("base_header.html")
	if err != nil {
		utils.Log.Fatal(err)
	}

	htmlCode += htmlHeader

	for _, item := range items {
		screenshotFileName := "./screenshots/" + GetScreenshotFileName(item.Url)
		modalName := strings.Split(GetScreenshotFileName(item.Url), ".")[0]
		//httpFileName := "./html/" + GetHTMLFileName(item.Url)

		// Modal Screenshot
		htmlCode += fmt.Sprintf("<div class=\"modal fade\" id=\"%sscreenshot\" tabindex=\"-1\" aria-labelledby=\"%sscreenshot\" aria-hidden=\"true\">\n", modalName, modalName)
		htmlCode += " <div class=\"modal-dialog modal-lg\">\n"
		htmlCode += "  <div class=\"modal-content\">\n"
		htmlCode += "   <div class=\"modal-header\">\n"
		htmlCode += fmt.Sprintf("<h5 class=\"modal-title text-center\">%s</h5>\n", item.Url)
		htmlCode += "   </div>\n"
		htmlCode += "   <div class=\"modal-body\">\n"
		htmlCode += fmt.Sprintf("    <img class=\"bd-placeholder-img card-img-top\" width=\"100%%\" height=\"100%%\" focusable=\"false\" src=\"%s\" />\n", screenshotFileName)
		htmlCode += "   </div>\n"
		htmlCode += "   <div class=\"modal-footer\">\n"
		htmlCode += fmt.Sprintf("      <a style=\"width: 100%%;\" href=\"%s\" target=\"_blank\" rel=\"noopener noreferrer\"><button type=\"button\" class=\"btn btn-sm btn-outline-secondary\" style=\"width: 100%%;\">Visit</button></a>\n", item.Url)
		htmlCode += "   </div>\n"
		htmlCode += "  </div>\n"
		htmlCode += " </div>\n"
		htmlCode += "</div>\n"

		// Modal Request
		htmlCode += fmt.Sprintf("<div class=\"modal fade\" id=\"%shtml\" tabindex=\"-1\" aria-labelledby=\"%shtml\" aria-hidden=\"true\">\n", modalName, modalName)
		htmlCode += " <div class=\"modal-dialog modal-lg\">\n"
		htmlCode += "  <div class=\"modal-content\">\n"
		htmlCode += "   <div class=\"modal-header\">\n"
		htmlCode += "    <h5 class=\"modal-title text-center\">Requests</h5>\n"
		htmlCode += "   </div>\n"
		htmlCode += "   <div class=\"modal-body\">\n"
		htmlCode += "    <pre>\n"

		// Headers
		for _, header := range item.ResponseHeaders {
			htmlCode += fmt.Sprintf("%s\n", header)
		}
		htmlCode += "    </pre>\n"

		htmlCode += "   </div>\n"
		htmlCode += "  </div>\n"
		htmlCode += " </div>\n"
		htmlCode += "</div>\n"

		// Card
		htmlCode += "<div class=\"col-md-4\">\n"
		htmlCode += " <div class=\"card mb-4 shadow-sm\">\n"

		// Header
		htmlCode += "  <div class=\"card-header text-center\">\n"
		htmlCode += fmt.Sprintf("      %s\n", item.Url)
		htmlCode += "  </div>\n"

		// Screenshot
		htmlCode += fmt.Sprintf("  <img class=\"bd-placeholder-img card-img-top\" width=\"100%%\" height=\"225\" focusable=\"false\" src=\"%s\" data-bs-toggle=\"modal\" data-bs-target=\"#%sscreenshot\"/>\n", screenshotFileName, modalName)

		// Card body
		htmlCode += "  <div class=\"card-body\">\n"

		// Response
		htmlCode += "   <hr>\n"
		htmlCode += "   <div class=\"btn-toolbar justify-content-between\" role=\"toolbar\">\n"
		htmlCode += " 	 <div class=\"btn-group me-2\" role=\"group\">\n"
		htmlCode += fmt.Sprintf("    <button type=\"button\" class=\"btn btn-primary\" style=\"user-select: all;\" disabled>%s %s</button>\n", item.ResponseStatus, item.ResponseStatusText)
		htmlCode += "    </div>\n"
		htmlCode += "    <div class=\"btn-group me-2\" role=\"group\">\n"
		htmlCode += fmt.Sprintf("    <button type=\"button\" class=\"btn btn-primary\" style=\"user-select: all;\" disabled>%s</button>\n", item.ResponseProtocol)
		htmlCode += "    </div>\n"
		htmlCode += " 	 <div class=\"btn-group me-2\" role=\"group\">\n"
		htmlCode += fmt.Sprintf("      <button type=\"button\" class=\"btn btn-info\" style=\"user-select: all;\" disabled>%s:%s</button>\n", item.RemoteIP, item.RemotePort)
		htmlCode += " 	 </div>\n"
		htmlCode += "   </div>\n"
		htmlCode += "   </br>\n"

		// Response Security info
		htmlCode += "   <hr>\n"
		htmlCode += " 	  <p>Certificate Issuer : <p>\n"
		htmlCode += " 	 <ul>\n"
		htmlCode += fmt.Sprintf("    <li>%s</li>\n", item.ResponseIssuer)
		htmlCode += "    </ul>\n"

		// Response Certificate subject
		htmlCode += "   <hr>\n"
		htmlCode += " 	  <p>Certificate Subject : <p>\n"
		htmlCode += " 	 <ul>\n"
		htmlCode += fmt.Sprintf("    <li>%s</li>\n", item.ResponseCertificateSubject)
		htmlCode += "    </ul>\n"

		// Response Certificate SanList
		htmlCode += "   <hr>\n"
		htmlCode += " 	  <p>Certificate Alternate Name : <p>\n"
		htmlCode += " 	 <ul>\n"
		for _, san := range item.ResponseCertificateSanList {
			htmlCode += fmt.Sprintf("    <li>%s</li>\n", san)
		}
		htmlCode += "    </ul>\n"
		htmlCode += "   <hr>\n"

		// Headers
		htmlCode += "   <div>\n"
		htmlCode += fmt.Sprintf("      <button type=\"button\" class=\"btn btn-outline-secondary\" data-bs-toggle=\"modal\" data-bs-target=\"#%shtml\" style=\"width: 100%%;\" >Headers</button>\n", modalName)
		htmlCode += "   </div>\n"
		htmlCode += "   <br>\n"

		// Visit
		htmlCode += "   <div>\n"
		htmlCode += fmt.Sprintf("      <a style=\"width: 100%%;\" href=\"%s\" target=\"_blank\" rel=\"noopener noreferrer\"><button type=\"button\" class=\"btn btn-sm btn-warning\" style=\"width: 100%%;\">Visit</button></a>\n", item.Url)
		htmlCode += "   </div>\n"

		// Close the card
		htmlCode += "  </div>\n"
		htmlCode += " </div>\n"
		htmlCode += "</div>\n"
		htmlCode += "\n\n"
	}

	htmlFooter, err := box.FindString("base_footer.html")
	if err != nil {
		utils.Log.Fatal(err)
	}

	htmlCode += htmlFooter

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
