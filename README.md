# EAS Golang SDK

# Golang SDK调用接口说明
|类|主要接口|描述|
|-----|------|------|
|PredictClient|NewPredictClient(endpoint, service_name) *PredictClient|PredictClient类构造函数，endpoint是服务端的endpoint地址，对于普通服务设置为默认网关endpoint；service_name为服务名字；两个参数不可缺失。|

# 程序示例


## 字符串输入输出程序示例

对于自定义Processor用户而言，通常采用字符串进行服务的输入输出调用(如pmml模型服务的调用)，具体的demo程序如下：

```go
package main

import (
        "fmt"
        "github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {
        client := eas.NewPredictClientWithConns("1828488879222746.cn-shanghai.pai-eas.aliyuncs.com", "scorecard_pmml_example", 10)
        client.SetToken("YWFlMDYyZDNmNTc3M2I3MzMwYmY0MmYwM2Y2MTYxMTY4NzBkNzdjOQ==")
        client.Init()
        req := "[{\"fea1\": 1, \"fea2\": 2}]"
        for i := 0; i < 1000; i++ {
                resp, err := client.StringPredict(req)
                if err != nil {
                        fmt.Printf("failed to predict: %v\n", err.Error())
                } else {
                        fmt.Printf("%v\n", resp)
                }
        }
}
```

## TensorFlow输入输出请求

Tensorflow用户可以使用TFRequest与TFResponse作为数据的输入输出格式，具体demo示例如下：

```go
package main

import (
        "fmt"
        "github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {
        cli := eas.NewPredictClient("1828488879222746.cn-shanghai.pai-eas.aliyuncs.com", "mnist_saved_model_example")
        cli.SetToken("YTg2ZjE0ZjM4ZmE3OTc0NzYxZDMyNmYzMTJjZTQ1YmU0N2FjMTAyMA==")
        cli.Init()

        tfreq := eas.TFRequest{}
        tfreq.SetSignatureName("predict_images")
        tfreq.AddFeedFloat32("images", []int64{1, 784}, make([]float32, 784))

        for i := 0; i < 1000; i++ {
                resp, err := cli.TFPredict(tfreq)
                if err != nil {
                        fmt.Printf("failed to predict: %v", err)
                } else {
                        fmt.Printf("%v\n", resp)
                }
        }
}
```


## PyTorch输入输出程序示例

PyTorch用户可以使用TorchRequest与TorchResponse作为数据的输入输出格式，具体demo示例如下：

```go
package main

import (
        "fmt"
        "github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {
        cli := eas.NewPredictClient("1828488879222746.cn-shanghai.pai-eas.aliyuncs.com", "torch_example")
        cli.SetToken("YTg2ZjE0ZjM4ZmE3OTc0NzYxZDMyNmYzMTJjZTQ1YmU0N2FjMTAyMA==")
        cli.Init()

        req := eas.TorchRequest{}
        req.AddFeedFloat32(0, eas.TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))
        req.AddFetch(0)

        for i := 0; i < 1000; i++ {
                resp, err := cli.TorchPredict(req)
                if err != nil {
                        fmt.Printf("failed to predict: %v\n", err)
                } else {
                        fmt.Printf("%v\n", resp)
                }
        }
}
```

## 通过VPC网络直连的方式调用服务

网络直连方式仅支持部署在EAS公共云控制台中购买专用资源组的服务，且需要在控制台上为该资源组与用户指定的vswitch打通网络后才可使用。调用方法与普通调用方式相比，增加一句 client.SetEndpointType(eas.EndpointTypeDirect) 即可，非常适合大流量高并发的服务。

```go
package main

import (
        "fmt"
        "github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {
        client := eas.NewPredictClientWithConns("1828488879222746.cn-shanghai.pai-eas.aliyuncs.com", "scorecard_pmml_example", 10)
        client.SetToken("YWFlMDYyZDNmNTc3M2I3MzMwYmY0MmYwM2Y2MTYxMTY4NzBkNzdjOQ==")
	client.SetEndpointType(eas.EndpointTypeDirect)
        client.Init()
        req := "[{\"fea1\": 1, \"fea2\": 2}]"
        for i := 0; i < 1000; i++ {
                resp, err := client.StringPredict(req)
                if err != nil {
                        fmt.Printf("failed to predict: %v\n", err.Error())
                } else {
                        fmt.Printf("%v\n", resp)
                }
        }
}
```

## 客户端连接参数的设置

可以通过设置http.Transport属性来设置请求客户端的连接参数，示例代码如下:

```go
package main

import (
        "fmt"
        "github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {
        client := eas.NewPredictClient("pai-eas-vpc.cn-shanghai.aliyuncs.com", "network_test")
        client.SetToken("MDAwZDQ3NjE3OThhOTI4ODFmMjJiYzE0MDk1NWRkOGI1MmVhMGI0Yw==")
        client.SetEndpointType(eas.EndpointTypeDirect)
        client.setHttpTransport(&http.Transport{
                MaxConnsPerHost:       300,
                TLSHandshakeTimeout:   100 * time.Millisecond,
                ResponseHeaderTimeout: 200 * time.Millisecond,
                ExpectContinueTimeout: 200 * time.Millisecond,
        })
}
```
