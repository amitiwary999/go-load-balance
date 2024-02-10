package loadbalanceroundrobin

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
)

type roundRobin struct {
	counter uint64
}

func NewRoundRobin() *roundRobin {
	return &roundRobin{
		counter: 0,
	}
}

func (r roundRobin) reverseProxy(w http.ResponseWriter, req *http.Request) error {
	counter := atomic.AddUint64(&r.counter, 1)
	url := URLs[counter%uint64(len(URLs))]
	completeUrl := fmt.Sprintf("%s%s", url, req.RequestURI)
	proxyReq, err := http.NewRequest(req.Method, completeUrl, req.Body)
	if err != nil {
		return errors.New("failed to create request")
	}
	proxyReq.Header = req.Header
	return nil
}

func (r roundRobin) RoundRobinHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		r.reverseProxy(w, req)
	}
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.reverseProxy(w, req)
}
