package loadbalancerrandom

import (
	lb "balanceload/load-balancer"
	proxy "balanceload/load-balancer/proxy"
	"math/rand"
	"net/http"
)

type random struct {
	urls []proxy.IProxy
}

func NewRandom(config *lb.Config, proxyFunc proxy.ProxyFunc) *random {
	var urls []proxy.IProxy
	for _, b := range config.Backends {
		urls = append(urls, proxyFunc(b.URL))
	}
	return &random{
		urls: urls,
	}
}

func (r *random) serve(w http.ResponseWriter, req *http.Request) {
	r.urls[rand.Intn(len(r.urls))].ReverseProxy(w, req)
}

func (r *random) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.serve(w, req)
}
