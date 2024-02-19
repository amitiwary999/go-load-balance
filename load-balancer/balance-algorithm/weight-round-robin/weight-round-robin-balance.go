package loadbalancerweightroundrobin

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"net/http"
	"sync/atomic"
)

type weightRoundRobin struct {
	counter uint64
	urls    []proxy.IProxy
}

func NewWeightRoundRobin(config *lb.Config, proxyFunc proxy.ProxyFunc) *weightRoundRobin {
	var urls []proxy.IProxy

	for _, b := range config.Backends {
		for i := 0; i < int(b.Weight); i++ {
			urls = append(urls, proxyFunc(b.URL))
		}
	}

	return &weightRoundRobin{
		urls: urls,
	}
}

func (r *weightRoundRobin) serve(w http.ResponseWriter, req *http.Request) {
	counter := atomic.AddUint64(&r.counter, 1)
	r.urls[counter%uint64(len(r.urls))].ReverseProxy(w, req)
}

func (r *weightRoundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}
