package basicauth

import (
	"fmt"
	"net/http"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
)

type BasicAuthMiddleware struct {
	Realm string
}

func (n BasicAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authorization := r.Header.Get("Authorization")

	if authorization == "" {
		basicHeader := "Basic"
		if n.Realm != "" {
			basicHeader = fmt.Sprintf("Basic realm=\"%s\"", n.Realm)
		}
		w.Header().Set("WWW-Authenticate", basicHeader)
		w.WriteHeader(401)
		return
	}

	if username, password, ok := r.BasicAuth(); ok {
		logger.Printf("[AUTH-BASIC] [%s] [%s] [%s]\n", utils.GetSourceIP(r), username, password)
	}

	next(w, r)
}
