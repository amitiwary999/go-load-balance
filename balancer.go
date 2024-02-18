package main

import (
	lb "balanceload/load-balancer"
	random "balanceload/load-balancer/balance-algorithm/random"
	rr "balanceload/load-balancer/balance-algorithm/round-robin"
	wrr "balanceload/load-balancer/balance-algorithm/weight-round-robin"
	proxy "balanceload/load-balancer/proxy"
	"fmt"
	"net"
	"net/http"
)

func main() {
	var handler http.Handler
	if lb.ParseConfig().AlgoType == "w-round-robin" {
		handler = wrr.NewWeightRoundRobin(lb.ParseConfig())
	} else if lb.ParseConfig().AlgoType == "round-robin" {
		handler = rr.NewRoundRobin(lb.ParseConfig())
	} else if lb.ParseConfig().AlgoType == "random" {
		handler = random.NewRandom(lb.ParseConfig(), proxy.NewProxy)
	}
	s := &http.Server{
		Handler: handler,
	}
	l, err := net.Listen("tcp4", ":8000")
	if err != nil {
		fmt.Printf("listener failed %v", err)
	}
	s.Serve(l)
}
