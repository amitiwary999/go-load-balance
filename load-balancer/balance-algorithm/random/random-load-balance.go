package loadbalancerrandom

import (
	lb "balanceload/load-balancer"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

type random struct {
	counter uint64
	urls    []string
}

func NewRandom(config *lb.Config) *random {
	var urls []string
	for _, b := range config.Backends {
		urls = append(urls, b.URL)
	}
	return &random{
		urls: urls,
	}
}

func (r *random) reverseProxy(w http.ResponseWriter, req *http.Request) error {
	if len(r.urls) == 0 {
		return errors.New("no server url")
	}
	url := r.urls[rand.Intn(len(r.urls))]
	completeUrl := fmt.Sprintf("%s%s", url, req.RequestURI)
	proxyReq, err := http.NewRequest(req.Method, completeUrl, req.Body)
	if err != nil {
		r.serverError(w, err.Error())
		return errors.New("failed to create request")
	}
	proxyReq.Header = req.Header
	client := &http.Client{}
	resp, respErr := client.Do(proxyReq)
	if respErr != nil {
		r.serverError(w, respErr.Error())
		fmt.Printf("response error %s\n", respErr)
		return respErr
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
	return nil
}

func (r *random) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.reverseProxy(w, req)
}

func (r *random) serverError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err))
}
