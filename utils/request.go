package utils

import "net/http"

func GetSourceIP(r *http.Request) string {
	if ipSrc := r.Header.Get("X-Real-Ip"); ipSrc != "" {
		return ipSrc
	}
	if ipSrc := r.Header.Get("X-Forwarded-For"); ipSrc != "" {
		return ipSrc
	}
	return r.RemoteAddr
}
