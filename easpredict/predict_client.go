package easpredict

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	// EndpointTypeDefault = EndPoint Type Default
	EndpointTypeDefault = "DEFAULT"
	// EndpointTypeVipserver = EndPoint Type Vipserver
	EndpointTypeVipserver = "VIPSERVER"
	// EdnpointTypeDirect = EndPoint Type Direct
	EdnpointTypeDirect = "DIRECT"
)

// PredictError is a custom err type
type PredictError struct {
	ErrorCode int
	ErrorMsg  string
}

// Error for error interface
func (prederr *PredictError) Error() string {
	msg := fmt.Sprintf("PredictError: Code: %d, Message: %s", prederr.ErrorCode, prederr.ErrorMsg)
	// fmt.Fprintf(os.Stderr, msg)
	return msg
}

// NewPredictError constructs an error
func NewPredictError(code int, msg string) *PredictError {
	return &PredictError{
		ErrorCode: code,
		ErrorMsg:  msg,
	}
}

// StringRequest for request interface
type StringRequest struct {
	str string
}

// ToString for request interface
func (s StringRequest) ToString() (string, error) {
	return s.str, nil
}

// StringResponse for response interface
type StringResponse struct {
	str string
}

func (s *StringResponse) unmarshal(body []byte) error {
	s.str = string(body)
	return nil
}

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

// NewPredictClientWithConnsTimeout returns an instance of PredictClient
func NewPredictClientWithConnsTimeout(endpointName string, serviceName string, maxConnsPerhost int, tlsHandshakeTimeout int, responseHeaderTimeout int, expectContinueTimeout int) *PredictClient {
	return &PredictClient{
		endpointName: endpointName,
		serviceName:  serviceName,
		retryCount:   5,
		client: http.Client{
			Timeout: 5000 * time.Millisecond,
			Transport: &http.Transport{
				MaxConnsPerHost:       maxConnsPerhost,
				TLSHandshakeTimeout:   time.Duration(tlsHandshakeTimeout) * time.Millisecond,
				ResponseHeaderTimeout: time.Duration(responseHeaderTimeout) * time.Millisecond,
				ExpectContinueTimeout: time.Duration(expectContinueTimeout) * time.Millisecond,
			},
		},
	}
}

// Init initialize client
func (p *PredictClient) Init() error {
	switch p.endpointType {
	case "":
		p.endpoint = newGatewayEndpoint(p.endpointName)
	case EndpointTypeDefault:
		p.endpoint = newGatewayEndpoint(p.endpointName)
	case EndpointTypeVipserver:
		p.endpoint = newVipServerEndpoint(p.endpointName)
	case EdnpointTypeDirect:
		p.endpoint = newCacheServerEndpoint(p.endpointName, p.serviceName)
	default:
		return NewPredictError(500, "Unsupported endpoint type: "+p.endpointType)
	}
	go p.syncHandler()
	return nil
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
	endName := p.endpoint.Get()
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
				return nil, NewPredictError(-1, err.Error())
			}
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			if i != p.retryCount-1 {
				continue
			}
			return body, NewPredictError(resp.StatusCode, resp.Status+":"+string(body))
		}
		return body, nil
	}
	return []byte{}, nil
}

// RequestIF interface
type RequestIF interface {
	ToString() (string, error)
}

// ResponseIF interface
type ResponseIF interface {
	unmarshal(body []byte) error
}

// Predict for request
func (p *PredictClient) Predict(request RequestIF) (ResponseIF, error) {
	req, err2 := request.ToString()
	if err2 != nil {
		return nil, err2
	}
	body, err := p.predict([]byte(req))
	if err != nil {
		return nil, err
	}

	switch request.(type) {
	case StringRequest:
		resp := StringResponse{}
		unmarshalerr := resp.unmarshal(body)
		return &resp, unmarshalerr
	case TFRequest:
		resp := TFResponse{}
		unmarshalerr := resp.unmarshal(body)
		return &resp, unmarshalerr
	case TorchRequest:
		resp := TorchResponse{}
		unmarshalerr := resp.unmarshal(body)
		return &resp, unmarshalerr
	default:
		return nil, NewPredictError(-1, "Unknown request type, currently support StringRequest, TFRequest and TorchRequest.")
	}
}

// StringPredict function send input data and return predicted result
func (p *PredictClient) StringPredict(str string) (string, error) {
	body, err := p.predict([]byte(str))
	return string(body), err
}

// TorchPredict function send input data and return PyTorch predicted result
func (p *PredictClient) TorchPredict(request TorchRequest) (TorchResponse, error) {
	resp, err := p.Predict(request)
	return *resp.(*TorchResponse), err
}

// TFPredict function send input data and return TensorFlow predicted result
func (p *PredictClient) TFPredict(request TFRequest) (TFResponse, error) {
	resp, err := p.Predict(request)
	return *resp.(*TFResponse), err
}
