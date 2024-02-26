package loadbalancer

import (
	"net/http"
)

var HopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func IsServerAlive(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}
