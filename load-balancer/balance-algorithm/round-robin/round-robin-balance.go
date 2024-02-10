package loadbalanceroundrobin

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type roundRobin struct {
	counter uint64
}

func NewRoundRobin() *roundRobin {
	return &roundRobin{}
}

func (r *roundRobin) reverseProxy(w http.ResponseWriter, req *http.Request) error {
	counter := atomic.AddUint64(&r.counter, 1)
	fmt.Printf("counter in req %v\n", counter)
	url := URLs[counter%uint64(len(URLs))]
	completeUrl := fmt.Sprintf("%s%s", url, req.RequestURI)
	proxyReq, err := http.NewRequest(req.Method, completeUrl, req.Body)
	if err != nil {
		return errors.New("failed to create request")
	}
	proxyReq.Header = req.Header
	client := &http.Client{}
	resp, respErr := client.Do(proxyReq)
	if respErr != nil {
		fmt.Printf("response error %s\n", respErr)
		return respErr
	}
	io.Copy(w, resp.Body)
	resp.Body.Close()
	return nil
}

func (r roundRobin) RoundRobinHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		r.reverseProxy(w, req)
	}
}

func (r *roundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.reverseProxy(w, req)
}
