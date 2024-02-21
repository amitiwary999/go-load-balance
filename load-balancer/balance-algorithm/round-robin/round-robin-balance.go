package loadbalanceroundrobin

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type backendServer struct {
	url           proxy.IProxy
	isServerAlive bool
}

type roundRobin struct {
	counter          uint64
	urls             []*backendServer
	backendServerMap map[string]*backendServer
}

func NewRoundRobin(config *lb.Config, proxyFunc proxy.ProxyFunc) *roundRobin {
	rr := &roundRobin{}
	var backendServers []*backendServer
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		b := &backendServer{url: proxyFunc(b.URL), isServerAlive: true}
		rr.backendServerMap[mapKey] = b
		backendServers = append(backendServers, b)
	}
	rr.urls = backendServers
	go rr.healthChecker(config)
	return rr
}

func (r *roundRobin) serve(w http.ResponseWriter, req *http.Request) {
	counter := atomic.AddUint64(&r.counter, 1)
	r.urls[counter%uint64(len(r.urls))].url.ReverseProxy(w, req)
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}

func (r *roundRobin) healthChecker(config *lb.Config) {
	for {
		time.Sleep(time.Duration(30 * time.Second))
		r.serverHealthCheck(config)
	}
}

func (r *roundRobin) serverHealthCheck(config *lb.Config) {
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		ok := lb.IsServerAlive(b.URL)
		if !ok && r.backendServerMap[mapKey].isServerAlive {
			r.backendServerMap[mapKey].isServerAlive = false
		}
	}
}
