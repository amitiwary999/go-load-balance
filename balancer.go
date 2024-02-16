package main

import (
	lb "balanceload/load-balancer"
	rr "balanceload/load-balancer/balance-algorithm/round-robin"
	"fmt"
	"net"
	"net/http"
)

func main() {
	s := &http.Server{
		Handler: rr.NewRoundRobin(lb.ParseConfig()),
	}
	l, err := net.Listen("tcp4", ":8000")
	if err != nil {
		fmt.Printf("listener failed %v", err)
	}
	s.Serve(l)
}
