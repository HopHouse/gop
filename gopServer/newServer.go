package gopserver

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hophouse/gop/utils/logger"
	"github.com/urfave/negroni"
)

type NewServer struct {
	Server Server
}

func (s NewServer) GetCertSubject() string {
	return s.Server.GetCertSubject()
}

func (s NewServer) GetServer(r *mux.Router, n *negroni.Negroni) (*http.Server, error) {
	addr := fmt.Sprintf("%s:%s", s.Server.Host, s.Server.Port)
	logger.Printf("[+] Serve file to: %s://%s\n", s.Server.Scheme, addr)

	n.UseHandler(r)

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	return &server, nil
}

func (s NewServer) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	})

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	return r
}

func (s NewServer) CreateMiddleware() *negroni.Negroni {
	return s.Server.CreateMiddleware()
}
