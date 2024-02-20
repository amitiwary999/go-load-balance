package loadbalanceroundrobin

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"net/http"
	"sync/atomic"
)

type backendServer struct {
	url           proxy.IProxy
	isServerAlive bool
}

type roundRobin struct {
	counter uint64
	urls    []*backendServer
}

func NewRoundRobin(config *lb.Config, proxyFunc proxy.ProxyFunc) *roundRobin {
	var backendServers []*backendServer
	for _, b := range config.Backends {
		backendServers = append(backendServers, &backendServer{url: proxyFunc(b.URL), isServerAlive: true})
	}
	return &roundRobin{
		urls: backendServers,
	}
}

func (r *roundRobin) serve(w http.ResponseWriter, req *http.Request) {
	counter := atomic.AddUint64(&r.counter, 1)
	r.urls[counter%uint64(len(r.urls))].url.ReverseProxy(w, req)
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}
