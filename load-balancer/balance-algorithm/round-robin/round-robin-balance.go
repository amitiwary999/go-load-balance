package loadbalanceroundrobin

import (
	lb "balanceload/load-balancer"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type roundRobin struct {
	counter uint64
	urls    []string
}

func NewRoundRobin(config *lb.Config) *roundRobin {
	var urls []string
	for _, b := range config.Backends {
		urls = append(urls, b.URL)
	}
	return &roundRobin{
		urls: urls,
	}
}

func (r *roundRobin) reverseProxy(w http.ResponseWriter, req *http.Request) error {
	counter := atomic.AddUint64(&r.counter, 1)
	url := r.urls[counter%uint64(len(r.urls))]
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

func (r roundRobin) RoundRobinHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		r.reverseProxy(w, req)
	}
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.reverseProxy(w, req)
}

func (r *roundRobin) serverError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err))
}
