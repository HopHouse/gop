package gopserver

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
)

type logMiddleware struct{}

func (l logMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("[%s] [%s] %s %s\n%s\n", time.Now().Format("2006.01.02 15:04:05"), utils.GetSourceIP(r), r.Method, r.URL.String(), string(reqDump))

	next(w, r)
}

type XRealIPMiddleware struct{}

func (n XRealIPMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	remoteIP := strings.Split(r.RemoteAddr, ":")[0]
	r.Header.Add("X-Real-IP-Full", r.RemoteAddr)
	r.Header.Add("X-Real-IP", remoteIP)
	next(w, r)
}
