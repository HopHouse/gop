package gopstaticcrawler

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/gookit/color"
	"github.com/hophouse/gop/utils/logger"
)

type Ressource struct {
	Url      string
	Type     string
	Secure   bool
	Internal bool
	sync.RWMutex
}

func CreateRessource(urlReference string, script string, kind string) (isInternal bool, ressource Ressource) {
	ressource.Internal = false
	ressource.Url = script
	ressource.Type = kind

	scriptUrl, err := url.Parse(script)
	if err != nil {
		logger.Println(err)
	}

	// Define if it is secure
	if scriptUrl.Scheme == "https" {
		ressource.Secure = true
	} else {
		ressource.Secure = false
	}

	// Define if its an internal or external script
	urlReferenceUrl, err := url.Parse(urlReference)
	if err != nil {
		logger.Println(err)
	}

	if scriptUrl.Host == urlReferenceUrl.Host || scriptUrl.Host == "" {
		ressource.Internal = true
	}
	return ressource.Internal, ressource
}

func (ressource Ressource) equal(newRessource Ressource) bool {
	ressource.RLock()
	if ressource.Url == newRessource.Url {
		if ressource.Type == newRessource.Type {
			if ressource.Secure == newRessource.Secure {
				ressource.RUnlock()
				return true
			}
		}
	}
	ressource.RUnlock()
	return false
}

func AddRessourceIfDoNotExists(ressources *[]Ressource, ressource Ressource) bool {
	for _, item := range *ressources {
		if added := item.equal(ressource); added == true {
			//logger.Println("[-] Ressource already present ", ressource.Url)
			return false
		}
	}

	*ressources = append(*ressources, ressource)
	return true
}

func (ressource Ressource) ressourceString() string {
	var result string
	var secureString string
	if ressource.Secure == true {
		secureString = color.FgGreen.Render("HTTPS")
	} else {
		secureString = color.FgRed.Render("HTTP")
	}

	result = fmt.Sprintf("[%v] [%s] %s", secureString, ressource.Type, ressource.Url)
	return result
}

func ressourceListString(ressources []Ressource) []string {
	result := make([]string, 0)
	for _, s := range ressources {
		result = append(result, s.ressourceString())
	}
	return result
}

func PrintRessourceList(ressources_string []Ressource) {
	tmp := ressourceListString(ressources_string)
	for _, s := range tmp {
		logger.Println("  - ", s)
	}
}

func (ressource Ressource) ressourceStringReport() string {
	var result string
	if ressource.Secure == true {
		result = fmt.Sprintf("\\rowcolor{gristableau} %s & %s & \\textbf{\\color{Green}HTTPS}\\\\ \n", ressource.Url, ressource.Type)
	} else {
		result = fmt.Sprintf("\\rowcolor{gristableau} %s & %s & \\textbf{\\color{output.Red}HTTP}\\\\ \n", ressource.Url, ressource.Type)
	}
	return result
}

func ressourceListStringReport(ressources []Ressource) []string {
	result := make([]string, 0)
	result = append(result, "\\begin{center}\n")
	result = append(result, "\\begin{tabular}{|c|c|c|}\n")
	result = append(result, "\\hline\n")
	for _, s := range ressources {
		if s.Type != "link" {
			result = append(result, s.ressourceStringReport())
			result = append(result, "\\hline\n")
		}
	}
	result = append(result, "\\end{center}\n")
	result = append(result, "\\end{tabular}\n")
	return result
}

func WriteRessourceListReport(ressources_string []Ressource) {
	tmp := ressourceListStringReport(ressources_string)
	f, errorFile := os.OpenFile("external_ressources.tex", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errorFile != nil {
		logger.Fatalf("error opening file: %v", errorFile)
	}
	defer f.Close()

	for _, s := range tmp {
		_, error := f.WriteString(s)
		if error != nil {
			logger.Println(error)
		}
	}
	f.Sync()
}
