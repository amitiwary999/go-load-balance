package loadbalanceroundrobin

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"net/http"
	"sync/atomic"
)

type roundRobin struct {
	counter uint64
	urls    []proxy.IProxy
}

func NewRoundRobin(config *lb.Config, proxyFunc proxy.ProxyFunc) *roundRobin {
	var urls []proxy.IProxy
	for _, b := range config.Backends {
		urls = append(urls, proxyFunc(b.URL))
	}
	return &roundRobin{
		urls: urls,
	}
}

func (r *roundRobin) serve(w http.ResponseWriter, req *http.Request) {
	counter := atomic.AddUint64(&r.counter, 1)
	r.urls[counter%uint64(len(r.urls))].ReverseProxy(w, req)
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}
