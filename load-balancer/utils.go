package loadbalancer

import "net/http"

func IsServerAlive(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}
