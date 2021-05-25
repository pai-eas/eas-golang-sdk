# EAS Golang SDK

# Golang SDK调用接口说明
|类|主要接口|描述|
|-----|------|------|
|PredictClient|NewPredictClient(endpoint, service_name) *PredictClient|PredictClient类构造函数，endpoint是服务端的endpoint地址，对于普通服务设置为默认网关endpoint；service_name为服务名字；两个参数不可缺失。|
||SetEndpoint(endpointName)|设置服务的endpoint，endpoint的说明见上述构造函数|
||SetServiceName(serviceName)|设置请求的服务名字|
||SetEndpointType(endpointType)|设置服务端的网关类型，支持默认网关("DEFAULT"或不设置），"VIPSERVER"，"DIRECT"，默认值为空|
||SetToken(token)|设置服务访问的token|
||SetHttpTransport(*transport)|设置http客户端的Transport属性|
||SetRetryCount(max_retry_count)|设置请求失败重试次数，默认为5；该参数非常重要，对于服务端进程异常或机器异常或网关长连接断开等情况带来的个别请求失败，均需由客户端来重试解决，请勿将其设置为0|
||SetTimeout(timeout)|设置请求的超时时间，单位为ms，默认为5000|
||Init() |对PredictClient对象进行初始化，在上述设置参数的函数执行完成后，**需要调用Init()函数才会生效**|
||Predict(Request)|向在线预测服务提交一个预测请求，request对象是interface(StringRequest, TFRequest,TorchRequest)，返回为Response interface(StringResponse, TFResponse,TorchResponse)|
||StringPredict(string)|向在线预测服务提交一个预测请求，request对象是string，返回也为string|
||TorchPredict(TorchRequest)|向在线预测服务提交一个预测请求，request对象是TorchRequest类，返回为对应的TorchResponse|
||TFPredict(TFRequest)|向在线预测服务提交一个预测请求，request对象是TFRequest类，返回为对应的TFResponse|
|StringRequest|StringRequest{string("")}|TFRequest类构建函数，将string转换为StringRequest以调用Predict方法|
|TFRequest|TFRequest(signature_name)|TFRequest类构建函数，输入为要请求模型的signature_name|
||AddFeed(?)(inputName string, shape []int64{}, content []?)|请求Tensorflow的在线预测服务模型时，设置需要输入的Tensor，inputName表示输入Tensor的别名，shape表示输入Tensor的TensorShape，content表示输入的Tensor的内容（一维数组展开表示），支持的类型包括Int32，Int64，Float32，Float64，String，Bool，函数名与具体类型相关，如AddFeedInt32()，若需要其它数据类型，可参考代码自行通过pb格式构造。 |
||AddFetch(outputName)|请求Tensorflow的在线预测服务模型时，设置需要输出的Tensor的别名，对于savedmodel模型该参数可选，若不设置，则输出所有的outputs，对于frozen model该参数必选|
|TFResponse|GetTensorShape(outputName)|获得别名为ouputname的输出Tensor的TensorShape|
||Get(?)Val(outputName)|获取输出的tensor的数据向量，输出结果以一维数组的形式保存，可配套使用GetTensorShape()获取对应的tensor的shape，将其还原成所需的多维tensor, 其中类型可选Float, Double, Int, Int64, String, Bool，函数名与具体类型相关，如GetFloatVal()|
|TorchRequest|TorchRequest()|TFRequest类构建方法|
||AddFeed(?)(index, shape []int64{}, content []?)|请求PyTorch的在线预测服务模型时，设置需要输入的Tensor，index表示要输入的tensor的下标，shape表示输入Tensor的TensorShape，content表示输入Tensor的内容（一维数组展开表示）。支持的类型包括Int32，Int64，Float32，Float64，函数名与具体类型相关，如AddFeedInt32()，若需要其它数据类型，可参考代码自行通过pb格式构造。 |
||AddFetch(outputIndex)|请求PyTorch的在线预测服务模型时，设置需要输出的Tensor的index，可选，若不设置，则输出所有的outputs|
|TorchResponse|GetTensorShape(outputIndex)|获得下标outputIndex的输出Tensor的TensorShape|
||Get(?)Val(outputIndex)|获取输出的tensor的数据向量，输出结果以一维数组的形式保存，可配套使用GetTensorShape()获取对应的tensor的shape，将其还原成所需的多维tensor, 其中类型可选Float, Double, Int, Int64，函数名与具体类型相关，如GetFloatVal()|

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
        cli := eas.NewPredictClient("1828488879222746.cn-shanghai.pai-eas.aliyuncs.co", "pytorch_resnet_example")
        cli.SetTimeout(500)
        cli.SetToken("ZjdjZDg1NWVlMWI2NTU5YzJiMmY5ZmE5OTBmYzZkMjI0YjlmYWVlZg==")
        cli.Init()
        req := eas.TorchRequest{}
        req.AddFeedFloat32(0, []int64{1, 3, 224, 224}, make([]float32, 150528))
        req.AddFetch(0)
        for i := 0; i < 100000000; i++ {
                resp, err := cli.TorchPredict(req)
                if err != nil {
                        fmt.Printf("failed to predict: %v", err)
                } else {
                        fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
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
