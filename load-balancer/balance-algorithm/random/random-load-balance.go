package loadbalancerrandom

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type backendServer struct {
	url           proxy.IProxy
	isServerAlive bool
}

type random struct {
	urls             []*backendServer
	backendServerMap map[string]*backendServer
}

func NewRandom(config *lb.Config, proxyFunc proxy.ProxyFunc) *random {
	ran := &random{}
	ran.backendServerMap = make(map[string]*backendServer)
	var backendServers []*backendServer
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		b := &backendServer{url: proxyFunc(b.URL), isServerAlive: true}
		ran.backendServerMap[mapKey] = b
		backendServers = append(backendServers, b)
	}
	ran.urls = backendServers
	go ran.healthChecker(config)
	return ran
}

func (r *random) serve(w http.ResponseWriter, req *http.Request) {
	if len(r.urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("none of the server is live"))
		return
	}
	r.urls[rand.Intn(len(r.urls))].url.ReverseProxy(w, req)
}

func (r *random) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}

func (r *random) healthChecker(config *lb.Config) {
	for {
		time.Sleep(time.Duration(10 * time.Second))
		r.serverHealthCheck(config)
	}
}

func (r *random) serverHealthCheck(config *lb.Config) {
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		ok := lb.IsServerAlive(b.URL)
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
