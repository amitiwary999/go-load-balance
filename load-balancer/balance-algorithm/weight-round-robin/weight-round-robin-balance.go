package loadbalancerweightroundrobin

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

type weightRoundRobin struct {
	counter          uint64
	urls             []*backendServer
	backendServerMap map[string]*backendServer
}

func NewWeightRoundRobin(config *lb.Config, proxyFunc proxy.ProxyFunc) *weightRoundRobin {
	rr := &weightRoundRobin{}
	rr.backendServerMap = make(map[string]*backendServer)
	var backendServers []*backendServer
	for indx, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(indx)
		ba := &backendServer{url: proxyFunc(b.URL), isServerAlive: true}
		rr.backendServerMap[mapKey] = ba
		for i := 0; i < int(b.Weight); i++ {
			backendServers = append(backendServers, ba)
		}
	}
	rr.urls = backendServers
	go rr.healthChecker(config)
	return rr
}

func (r *weightRoundRobin) serve(w http.ResponseWriter, req *http.Request) {
	if len(r.urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("none of the server is live"))
		return
	}
	counter := atomic.AddUint64(&r.counter, 1)
	r.urls[counter%uint64(len(r.urls))].url.ReverseProxy(w, req)
}

func (r *weightRoundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}

func (r *weightRoundRobin) healthChecker(config *lb.Config) {
	for {
		time.Sleep(time.Duration(10 * time.Second))
		r.serverHealthCheck(config)
	}
}

func (r *weightRoundRobin) serverHealthCheck(config *lb.Config) {
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		ok := lb.IsServerAlive(b.URL)
		if !ok && r.backendServerMap[mapKey].isServerAlive {
			r.backendServerMap[mapKey].isServerAlive = false
			r.urls = resizeServer(r.backendServerMap[mapKey], r.urls, b.Weight)
		} else if ok && !r.backendServerMap[mapKey].isServerAlive {
			r.backendServerMap[mapKey].isServerAlive = true
			bs := r.backendServerMap[mapKey]
			for i = 0; i < int(b.Weight); i++ {
				r.urls = append(r.urls, bs)
			}
		}
	}
	if len(r.urls) <= 0 {
		panic("none of the server is live")
	}
}

func resizeServer(b *backendServer, bs []*backendServer, weight uint) []*backendServer {
	rs := bs
	start := 1
	end := 1
	for i, bsp := range bs {
		if bsp == b {
			start = i
			end = start + int(weight)
			break
		}
	}
	rs = append(rs[:start], rs[end:]...)
	return rs
}
