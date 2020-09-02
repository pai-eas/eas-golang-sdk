package easpredict

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"eas-golang-sdk/easpredict/tf_predict_protos"
	"eas-golang-sdk/easpredict/torch_predict_protos"

	"github.com/golang/protobuf/proto"
)

// PredictClient for accessing prediction service by creating a fixed size connection pool
// to perform the request through established persistent connections.
type PredictClient struct {
	retryCount         int
	maxConnectionCount int
	token              string
	endpoint           endpointIF
	endpointType       string
	endpointName       string
	serviceName        string
	stop               bool
	client             http.Client
}

// NewPredictClient returns an instance of PredictClient
func NewPredictClient(endpointName string, serviceName string) *PredictClient {
	return &PredictClient{
		endpointName: endpointName,
		serviceName:  serviceName,
		retryCount:   5,
		client: http.Client{
			Timeout: 5000 * time.Millisecond,
			Transport: &http.Transport{
				MaxConnsPerHost: 100,
			},
		},
	}
}

// NewPredictClientWithConns returns an instance of PredictClient
func NewPredictClientWithConns(endpointName string, serviceName string, maxConnsPerhost int) *PredictClient {
	return &PredictClient{
		endpointName: endpointName,
		serviceName:  serviceName,
		retryCount:   5,
		client: http.Client{
			Timeout: 5000 * time.Millisecond,
			Transport: &http.Transport{
				MaxConnsPerHost: maxConnsPerhost,
			},
		},
	}
}

// Init initialize client
func (p *PredictClient) Init() {
	if p.endpointType == "" || p.endpointType == "DEFAULT" {
		p.endpoint = newGatewayEndpoint(p.endpointName)
	} else if p.endpointType == "VIPSERVER" {
		p.endpoint = newVipServerEndpoint(p.endpointName)
	} else if p.endpointType == "DIRECT" {
		p.endpoint = newCacheServerEndpoint(p.endpointName, p.serviceName)
	} else {
		defer fmt.Println("Code: 500, Message: Unsupported endpoint type: ", p.endpointType)
	}
	go p.syncHandler()
}

// syncHandler sync endpoint with server
func (p *PredictClient) syncHandler() {
	for true {
		if p.stop {
			break
		}
		p.endpoint.sync()
		time.Sleep(3 * time.Second)
	}
}

// SetEndpoint for client
func (p *PredictClient) SetEndpoint(endpointName string) {
	p.endpointName = endpointName
}

// SetEndpointType for client
func (p *PredictClient) SetEndpointType(endpointType string) {
	p.endpointType = endpointType
}

// SetToken function sets token for client
func (p *PredictClient) SetToken(token string) {
	p.token = token
}

// SetRetryCount for client
func (p *PredictClient) SetRetryCount(cnt int) {
	p.retryCount = cnt
}

// SetTimeout for client
func (p *PredictClient) SetTimeout(timeout int) {
	p.client.Timeout = time.Duration(timeout) * time.Millisecond
}

// SetServiceName for client
func (p *PredictClient) SetServiceName(serviceName string) {
	p.serviceName = serviceName
}

// buildURI returns an url for request
func (p *PredictClient) buildURI() string {
	endName := p.endpointName
	endName = p.endpoint.Get()
	if len(p.serviceName) != 0 {
		if p.serviceName[len(p.serviceName)-1] == '/' {
			p.serviceName = p.serviceName[:len(p.serviceName)-1]
		}
	}
	return fmt.Sprintf("http://%s/api/predict/%s", endName, p.serviceName)
}

func (p *PredictClient) tryNext(url string) string {
	addr := url[strings.Index(url, "http://")+len("http://") : strings.Index(url, "/api/predict")]
	endName := p.endpoint.TryNext(addr)
	if len(p.serviceName) != 0 {
		if p.serviceName[len(p.serviceName)-1] == '/' {
			p.serviceName = p.serviceName[:len(p.serviceName)-1]
		}
	}
	return fmt.Sprintf("http://%s/api/predict/%s", endName, p.serviceName)
}

// predict function posts inputs rawData to server and get response as []byte{}
func (p *PredictClient) predict(rawData []byte) ([]byte, error) {
	url := p.buildURI()
	for i := 0; i < p.retryCount; i++ {
		if i != 0 {
			url = p.tryNext(url)
		}
		req, _ := http.NewRequest("POST", url, bytes.NewReader(rawData))
		req.Header.Set("Content-Type", "application/octet-stream")
		if p.token != "" {
			req.Header.Set("Authorization", p.token)
		}
		resp, err := p.client.Do(req)
		if err != nil {
			if i == p.retryCount-1 {
				fmt.Println("request error:", err)
				return nil, err
			}
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil || resp.StatusCode != 200 {
			if i != p.retryCount-1 {
				continue
			}
			fmt.Println("request error:", resp.Status, string(body))
			fmt.Println(err)
			return body, errors.New(resp.Status)
		}
		return body, nil
	}
	return []byte{}, nil
}

// StringPredict function send input data and return predicted result
func (p *PredictClient) StringPredict(str string) (string, error) {
	body, err := p.predict([]byte(str))
	return string(body), err
}

// TorchPredict function send input data and return PyTorch predicted result
func (p *PredictClient) TorchPredict(request TorchRequest) (TorchResponse, error) {
	reqdata, err := proto.Marshal(&request.RequestData)
	if err != nil {
		fmt.Println("Marshal error: ", err)
	}

	body, err := p.predict(reqdata)
	bd := &torch_predict_protos.PredictResponse{}
	proto.Unmarshal(body, bd)
	rsp := &TorchResponse{*bd}

	return *rsp, err
}

// TFPredict function send input data and return TensorFlow predicted result
func (p *PredictClient) TFPredict(request TFRequest) (TFResponse, error) {
	reqdata, err := proto.Marshal(&request.RequestData)
	if err != nil {
		fmt.Println("Marshal error: ", err)
	}

	body, err := p.predict(reqdata)
	bd := &tf_predict_protos.PredictResponse{}
	proto.Unmarshal(body, bd)
	rsp := &TFResponse{*bd}

	return *rsp, err
}
