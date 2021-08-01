package gopserver

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hophouse/gop/utils"
)

type logMiddleware struct{}

func (l logMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Printf("[%s] [%s] %s %s\n", time.Now().Format("2006.01.02 15:04:05"), utils.GetSourceIP(r), r.Method, r.URL.String())
	utils.Log.Printf("[%s] [%s] %s %s\n", time.Now().Format("2006.01.02 15:04:05"), utils.GetSourceIP(r), r.Method, r.URL.String())

	utils.Log.Printf("%s %s %s", r.Method, r.URL, r.Proto)
	for k, v := range r.Header {
		for _, vv := range v {
			utils.Log.Printf("%s: %s ", k, vv)
		}
	}
	next(w, r)
}

type XRealIPMiddleware struct{}

func (n XRealIPMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	remoteIP := strings.Split(r.RemoteAddr, ":")[0]
	r.Header.Add("X-Real-IP-Full", r.RemoteAddr)
	r.Header.Add("X-Real-IP", remoteIP)
	next(w, r)
}
