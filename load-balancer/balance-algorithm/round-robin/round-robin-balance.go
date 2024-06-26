package loadbalanceroundrobin

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	lb "github.com/amitiwary999/go-load-balance/load-balancer"
	proxy "github.com/amitiwary999/go-load-balance/load-balancer/proxy"
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
	rr.backendServerMap = make(map[string]*backendServer)
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
	if len(r.urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("none of the server is live"))
		return
	}
	counter := atomic.AddUint64(&r.counter, 1)
	r.urls[counter%uint64(len(r.urls))].url.ReverseProxy(w, req)
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}

func (r *roundRobin) healthChecker(config *lb.Config) {
	for {
		time.Sleep(time.Duration(10 * time.Second))
		r.serverHealthCheck(config)
	}
}

func (r *roundRobin) serverHealthCheck(config *lb.Config) {
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		healthUrl := fmt.Sprintf("%v%v", b.URL, b.Health)
		ok := lb.IsServerAlive(healthUrl)
		if !ok && r.backendServerMap[mapKey].isServerAlive {
			r.backendServerMap[mapKey].isServerAlive = false
			r.urls = resizeServer(r.backendServerMap[mapKey], r.urls)
		} else if ok && !r.backendServerMap[mapKey].isServerAlive {
			r.backendServerMap[mapKey].isServerAlive = true
			r.urls = append(r.urls, r.backendServerMap[mapKey])
		}
	}
	if len(r.urls) <= 0 {
		panic("none of the server is live")
	}
}

func resizeServer(b *backendServer, bs []*backendServer) []*backendServer {
	rs := bs
	for i, bsp := range bs {
		if bsp == b {
			rs = append(bs[:i], bs[i+1:]...)
			break
		}
	}
	return rs
}
