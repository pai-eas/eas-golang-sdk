package easpredict

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type endpointIF interface {
	sync()
	Get() string
	TryNext(addr string) string
}

type endpointVar struct {
	lock      sync.RWMutex
	endPoints map[string]int
	scheduler wrrscheduler
}

func newEndpoint() *endpointVar {
	sc := &wrrscheduler{inited: false}
	ep := make(map[string]int)
	return &endpointVar{
		lock:      sync.RWMutex{},
		endPoints: ep,
		scheduler: *sc,
	}
}

// setEndpoints for endpointVar
func (ep *endpointVar) setEndpoints(endpoints map[string]int) {
	ep.lock.Lock()
	ep.endPoints = endpoints
	ep.scheduler = wrrScheduler(ep.endPoints)
	ep.lock.Unlock()
}

// Get ip address and port wrr returned
func (ep *endpointVar) Get() string {
	for true {
		if ep.scheduler.inited {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	ep.lock.RLock()
	addr := ep.scheduler.getNext()
	ep.lock.RUnlock()
	return addr
}

func (ep *endpointVar) TryNext(addr string) string {
	if len(ep.endPoints) > 1 {
		newaddr := ep.Get()
		for addr == newaddr {
			newaddr = ep.Get()
		}
		return newaddr
	}
	return ep.Get()
}

type cacheServerEndpoint struct {
	endpointVar
	domain      string
	serviceName string
	client      http.Client
}

// newCacheServerEndpoint returns an instance of cacheServerEndpoint
func newCacheServerEndpoint(domain string, serviceName string) *cacheServerEndpoint {
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.Replace(domain, "https://", "", 1)
	if domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}

	return &cacheServerEndpoint{
		domain:      domain,
		serviceName: serviceName,
		client:      http.Client{},
	}
}

// sync with server, get server list and set endpoints
func (c *cacheServerEndpoint) sync() {
	c.domain = strings.Replace(c.domain, "http://", "", 1)
	c.domain = strings.Replace(c.domain, "https://", "", 1)
	url := fmt.Sprintf("http://%s/exported/apis/eas.alibaba-inc.k8s.io/v1/upstreams/%s", c.domain, c.serviceName)
	endpoints := make(map[string]int)
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Println("sync service endpoints error: ", resp.Status, string(body))
		return
	}
	defer resp.Body.Close()
	result := make(map[string]interface{})
	json.Unmarshal(body, &result)
	hosts := result["endpoints"].(map[string]interface{})["items"].([]interface{})
	for _, hostmap := range hosts {
		host := hostmap.(map[string]interface{})
		endpoints[host["ip"].(string)+":"+fmt.Sprint(host["port"].(float64))] = int(host["weight"].(float64))
	}
	c.setEndpoints(endpoints)
}

type gatewayEndpoint struct {
	domain string
}

// newGatewayEndpoint returns an instance of gatewayEndpoint
func newGatewayEndpoint(domain string) *gatewayEndpoint {
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.Replace(domain, "https://", "", 1)
	if domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}

	return &gatewayEndpoint{
		domain: domain,
	}
}

// rewrite Get() function
func (g *gatewayEndpoint) Get() string {
	return g.domain
}

// rewrite TryNext() function
func (g *gatewayEndpoint) TryNext(addr string) string {
	return g.domain
}

// sync for the interface
func (g *gatewayEndpoint) sync() {
	// do nothing for gatewayEndpoint
	return
}

type vipServerEndpoint struct {
	endpointVar
	domain string
	client http.Client
}

// newVipServerEndpoint returns an instance for vipServerEndpoint
func newVipServerEndpoint(domain string) *vipServerEndpoint {
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.Replace(domain, "https://", "", 1)
	if domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}

	return &vipServerEndpoint{
		endpointVar: *newEndpoint(),
		domain:      domain,
		client:      http.Client{},
	}
}

// getServer() gets a random server from serverlist
func (v *vipServerEndpoint) getServer() (string, error) {
	vipserend := "http://jmenv.tbsite.net:8080/vipserver/serverlist"
	req, _ := http.NewRequest("GET", vipserend, nil)
	resp, err := v.client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		return resp.Status, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("sync service endpoints error: ", resp.StatusCode, err)
		return resp.Status, err
	}
	serverList := strings.Split(strings.Trim(string(body[:]), " "), "\n")
	rand.Seed(time.Now().UTC().UnixNano())
	return serverList[rand.Intn(len(serverList)-1)], nil
}

// sync with server, get server list and set endpoints
func (v *vipServerEndpoint) sync() {
	server, err := v.getServer()
	if err != nil {
		fmt.Println("Get server lists error: ", err)
		return
	}
	url := fmt.Sprintf("http://%s/vipserver/api/srvIPXT?dom=%s&clusters=DEFAULT", server, v.domain)

	endpoints := make(map[string]int)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := v.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("sync service endpoints error: %s, %s\n", body, err)
		return
	}

	mp := make(map[string]interface{})
	json.Unmarshal(body, &mp)
	for _, hostmap := range mp["hosts"].([]interface{}) {
		host := hostmap.(map[string]interface{})
		if host["valid"].(bool) {
			endpoints[host["ip"].(string)+":"+fmt.Sprint(host["port"].(float64))] = int(host["weight"].(float64))
		}
	}

	v.setEndpoints(endpoints)
}
