package main

import (
	"fmt"
	"net"
	"net/http"

	lb "github.com/amitiwary999/go-load-balance/load-balancer"
	random "github.com/amitiwary999/go-load-balance/load-balancer/balance-algorithm/random"
	rr "github.com/amitiwary999/go-load-balance/load-balancer/balance-algorithm/round-robin"
	wrr "github.com/amitiwary999/go-load-balance/load-balancer/balance-algorithm/weight-round-robin"
	proxy "github.com/amitiwary999/go-load-balance/load-balancer/proxy"
)

func main() {
	fmt.Printf("server main %v\n", lb.ParseConfig().AlgoType)
	var handler http.Handler
	if lb.ParseConfig().AlgoType == "w-round-robin" {
		handler = wrr.NewWeightRoundRobin(lb.ParseConfig(), proxy.NewProxy)
	} else if lb.ParseConfig().AlgoType == "round-robin" {
		handler = rr.NewRoundRobin(lb.ParseConfig(), proxy.NewProxy)
	} else if lb.ParseConfig().AlgoType == "random" {
		handler = random.NewRandom(lb.ParseConfig(), proxy.NewProxy)
	}
	fmt.Printf("server setup")
	s := &http.Server{
		Handler: handler,
	}
	l, err := net.Listen("tcp4", ":8000")
	if err != nil {
		fmt.Printf("listener failed %v", err)
	} else {
		fmt.Printf("server connected")
	}
	s.Serve(l)
}
