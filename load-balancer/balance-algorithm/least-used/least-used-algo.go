package loadbalanceleastused

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"net/http"
	"strconv"
)

type backendServer struct {
	url           proxy.IProxy
	isServerAlive bool
}

type leastUsed struct {
	urls             []*backendServer
	backendServerMap map[string]*backendServer
}

func NewLeastUsed(config *lb.Config, proxyFunc proxy.ProxyFunc) *leastUsed {
	ran := &leastUsed{}
	ran.backendServerMap = make(map[string]*backendServer)
	var backendServers []*backendServer
	for i, b := range config.Backends {
		mapKey := b.URL + strconv.Itoa(i)
		b := &backendServer{url: proxyFunc(b.URL), isServerAlive: true}
		ran.backendServerMap[mapKey] = b
		backendServers = append(backendServers, b)
	}
	ran.urls = backendServers
	return ran
}

func (r *leastUsed) serve(w http.ResponseWriter, req *http.Request) {
	if len(r.urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("none of the server is live"))
		return
	}
	proxy := r.urls[0].url
	for _, server := range r.urls {
		if server.url.PendingRequests() < proxy.PendingRequests() {
			proxy = server.url
		}
	}
	proxy.ReverseProxy(w, req)
}

func (r *leastUsed) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}
