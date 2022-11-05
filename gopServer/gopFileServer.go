package gopserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

type FilseServer struct {
	Server    Server
	Directory string
}

func (fs FilseServer) GetCertSubject() string {
	return fs.Server.GetCertSubject()
}

func (fs FilseServer) GetServer(r *mux.Router, n *negroni.Negroni) (http.Server, error) {
	path, err := os.Getwd()
	if err != nil {
		return http.Server{}, err
	}

	if !strings.HasPrefix(fs.Directory, "/") && !strings.HasPrefix(fs.Directory, "C:\\") {
		fs.Directory = filepath.Join(path, fs.Directory)
	}

	addr := fmt.Sprintf("%s:%s", fs.Server.Host, fs.Server.Port)
	logger.Printf("[+] Serve file to: %s://%s for %s\n", fs.Server.Scheme, addr, fs.Directory)

	n.UseHandler(r)

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	return server, nil
}

func (fs FilseServer) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	fileServer := http.FileServer(http.Dir(fs.Directory))
	r.PathPrefix("/").Handler(fileServer)

	return r
}

func (fs FilseServer) CreateMiddleware() *negroni.Negroni {
	return fs.Server.CreateMiddleware()
}
