package loadbalancerproxy

import (
	lb "balanceload/load-balancer"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ProxyFunc func(url string) IProxy

type IProxy interface {
	ReverseProxy(w http.ResponseWriter, req *http.Request) error
	serverError(w http.ResponseWriter, err string)
}

type Proxy struct {
	backendUrl string
}

func NewProxy(url string) IProxy {
	return &Proxy{
		backendUrl: url,
	}
}

func (r *Proxy) serverError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err))
}

func (r *Proxy) ReverseProxy(w http.ResponseWriter, req *http.Request) error {
	completeUrl := fmt.Sprintf("%s%s", r.backendUrl, req.RequestURI)
	proxyReq, err := http.NewRequest(req.Method, completeUrl, req.Body)
	if err != nil {
		r.serverError(w, err.Error())
		return errors.New("failed to create request")
	}
	removeHopHeaders(req)
	proxyReq.Header = req.Header
	client := &http.Client{}
	resp, respErr := client.Do(proxyReq)
	if respErr != nil {
		r.serverError(w, respErr.Error())
		fmt.Printf("response error %s\n", respErr)
		return respErr
	}
	removeResHopHeader(resp)
	w.WriteHeader(resp.StatusCode)
	for h, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Set(h, v)
		}
	}
	io.Copy(w, resp.Body)
	resp.Body.Close()
	return nil
}

func removeHopHeaders(req *http.Request) {
	for _, h := range lb.HopHeaders {
		req.Header.Del(h)
	}
}

func removeResHopHeader(resp *http.Response) {
	for _, h := range lb.HopHeaders {
		resp.Header.Del(h)
	}
}
