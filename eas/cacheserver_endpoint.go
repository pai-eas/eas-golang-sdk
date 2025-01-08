package eas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type cacheServerEndpoint struct {
	baseEndpoint
	domain      string
	serviceName string
	internal    bool
	client      http.Client
}

type CacheServerOption func(c *cacheServerEndpoint)

func WithInternalDirect(internal bool) func(c *cacheServerEndpoint) {
	return func(c *cacheServerEndpoint) {
		c.internal = internal
	}
}

// newCacheServerEndpoint returns an instance of cacheServerEndpoint
func newCacheServerEndpoint(domain string, serviceName string, options ...CacheServerOption) *cacheServerEndpoint {
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.Replace(domain, "https://", "", 1)
	if len(domain) > 0 && domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}
	c := &cacheServerEndpoint{
		domain:      domain,
		serviceName: serviceName,
		internal:    false,
		client:      http.Client{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// sync synchronizes the service's endpoints from upstream cache server and replace the endpoints in memory
func (c *cacheServerEndpoint) Sync() {
	c.domain = strings.Replace(c.domain, "http://", "", 1)
	c.domain = strings.Replace(c.domain, "https://", "", 1)
	url := fmt.Sprintf("http://%s/exported/apis/eas.alibaba-inc.k8s.io/v1/upstreams/%s?internal=%v", c.domain, c.serviceName, c.internal)
	endpoints := make(map[string]int)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("failed to query %v: %v", url, err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read response body from %v: %v", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("failed to sync service endpoints from %v: %v, %v", url, resp.Status, string(body))
		return
	}
	result := make(map[string]interface{})
	json.Unmarshal(body, &result)
	hosts := result["endpoints"].(map[string]interface{})["items"].([]interface{})
	for _, hostmap := range hosts {
		host := hostmap.(map[string]interface{})
		name := fmt.Sprintf("%v:%v", host["ip"], host["port"])
		endpoints[name] = int(host["weight"].(float64))
	}
	c.setEndpoints(endpoints)
}
