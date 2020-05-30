package gopactivecrawler

import (
    "fmt"
    "strings"
	"os"
	"sync"

    "github.com/hophouse/gop/utils"
	"github.com/gookit/color"
)

type Ressource struct {
    Url string
    Type string
    Secure bool
	mutex sync.Mutex
}

func CreateRessource(urlReference string, script string, kind string) (isInternal bool, ressource Ressource) {
    isInternal = false
    ressource.Url = script
    ressource.Type = kind

    // Define if it is secure
    if strings.HasPrefix(script, "https") {
        ressource.Secure = true
    } else {
        ressource.Secure = false
    }

    // Define if its an internal or external script
    if strings.HasPrefix(script, urlReference) {
        isInternal = true
    }
    return isInternal, ressource
}

func (ressource Ressource) equal(newRessource Ressource) (bool) {
    if ressource.Url == newRessource.Url {
        if ressource.Type == newRessource.Type {
            if ressource.Secure == newRessource.Secure {
                return true
            }
        }
    }
    return false
}

func AddRessourceIfDoNotExists(ressources *[]Ressource, ressource Ressource) (bool) {
    for _,item := range *ressources {
        if added := item.equal(ressource); added == true {
            utils.Log.Println("[-] Ressource already present ", ressource.Url)
            return false
        }
    }
    ressource.mutex.Lock()
    *ressources = append(*ressources, ressource)
    ressource.mutex.Unlock()
    return true
}

func (ressource Ressource) ressourceString() string {
    var result string
    if ressource.Secure == true {
        result = fmt.Sprintf("%s - %s - %v", ressource.Url, ressource.Type, color.FgGreen.Render("HTTPS"))
    } else {
        result = fmt.Sprintf("%s - %s - %v", ressource.Url, ressource.Type, color.FgGreen.Render("HTTP"))
    }
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
        color.Println("  - ", s)
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
        if (s.Type != "link") {
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
    f, errorFile := os.OpenFile("external_ressources.tex", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if errorFile != nil {
        utils.Log.Fatalf("error opening file: %v", errorFile)
    }
    defer f.Close()

    for _, s := range tmp {
        _, error := f.WriteString(s)
        if (error != nil) {
            color.Println(error)
        }
    }
    f.Sync()
}
