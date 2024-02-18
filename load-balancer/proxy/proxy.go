package loadbalancerproxy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ProxyFunc func(url string) IProxy

type IProxy interface {
	reverseProxy(w http.ResponseWriter, req *http.Request, url string) error
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

func (r *Proxy) reverseProxy(w http.ResponseWriter, req *http.Request, url string) error {
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
