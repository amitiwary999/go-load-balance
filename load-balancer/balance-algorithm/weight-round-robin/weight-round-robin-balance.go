package loadbalancerweightroundrobin

import (
	lb "balanceload/load-balancer"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type WeightRoundRobin struct {
	counter uint64
	urls    []string
}

func NewWeightRoundRobin(config *lb.Config) *WeightRoundRobin {
	var urls []string

	for _, b := range config.Backends {
		for i := 0; i < int(b.Weight); i++ {
			urls = append(urls, b.URL)
		}
	}

	return &WeightRoundRobin{
		urls: urls,
	}
}

func (r *WeightRoundRobin) reverseProxy(w http.ResponseWriter, req *http.Request) error {
	counter := atomic.AddUint64(&r.counter, 1)
	url := r.urls[counter%uint64(len(r.urls))]
	completeUrl := fmt.Sprintf("%s%s", url, req.RequestURI)
	proxyReq, err := http.NewRequest(req.Method, completeUrl, req.Body)
	if err != nil {
		r.serverError(w, err.Error())
		return errors.New("failed to create request")
	}
	proxyReq.Header = req.Header
	client := &http.Client{}
	resp, respErr := client.Do(proxyReq)
	if respErr != nil {
		r.serverError(w, respErr.Error())
		fmt.Printf("response error %s\n", respErr)
		return respErr
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
	return nil
}

func (r *WeightRoundRobin) WRoundRobinHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		r.reverseProxy(w, req)
	}
}

func (r *WeightRoundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	r.reverseProxy(w, req)
}

func (r *WeightRoundRobin) serverError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err))
}
