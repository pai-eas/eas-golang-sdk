# EAS Golang SDK (BETA)

**请用户注意：本sdk为beta版本，可供调试和试用。线上请谨慎试用！**

**请用户注意：本sdk为beta版本，可供调试和试用。线上请谨慎试用！**

**请用户注意：本sdk为beta版本，可供调试和试用。线上请谨慎试用！**

总体参考了Python和Java SDK的设计和实现，定义predict_client做为请求客户端，内部方法实现了发送http请求，对于vipserver和direct的实现，在endpoint中定义相关类型，也参考了python的实现，wrr.go实现了roud robin算法。

对pytorch和TensorFlow请求的格式进行了包装，并按照相同风格定义方法，对直接请求的情况已跑通。


# Golang SDK调用接口说明

|类|主要接口|描述|
|-----|------|------|
|PredictClient|NewPredictClient(endpoint, service_name)|PredictClient类构造器，endpoint是服务端的endpoint地址，对于普通服务设置为默认网关endpoint；service_name为服务名字；两个参数不可缺失。|
||NewPredictClientWithConns(endpointName string, serviceName string, maxConnsPerhost int)|带有连接池配置的构造器，maxConnsPerhost是连接池中对每个host的最大连接数|
||NewPredictClientWithConnsTimeout(endpointName string, serviceName string, maxConnsPerhost int, tlsHandshakeTimeout int, responseHeaderTimeout int, expectContinueTimeout int)|带有连接池和链接超时配置的构造器，maxConnsPerhost是连接池中对每个host的最大连接数，tlsHandshakeTimeout，responseHeaderTimeout，expectContinueTimeout分别为建立链接超时时间、等待回应请求头的时间、回应请求头接受之后继续等待的时间，详细定义见net/http中Transport的对应定义|
||SetEndpoint(endpointName)|设置服务的endpoint，endpoint的说明见上述构造函数|
||SetServiceName(serviceName)|设置请求的服务名字|
||SetEndpointType(endpointType)|设置服务端的网关类型，支持默认网关("DEFAULT"或不设置），"VIPSERVER"，"DIRECT"，默认值为空|
||SetToken(token)|设置服务访问的token|
||SetRetryCount(max_retry_count)|设置请求失败重试次数，默认为5；该参数非常重要，对于服务端进程异常或机器异常或网关长连接断开等情况带来的个别请求失败，均需由客户端来重试解决，请勿将其设置为0|
||SetTimeout(timeout)|设置请求的超时时间，单位为ms，默认为5000|
||Init() |对PredictClient对象进行初始化，在上述设置参数的函数执行完成后，**需要调用Init()函数才会生效**|
||Predict(RequestIF interface)|向在线预测服务提交一个预测请求，request对象是interface(StringRequest, TFRequest,TorchRequest)，返回为Response interface(StringResponse, TFResponse,TorchResponse)|
||StringPredict(string)|向在线预测服务提交一个预测请求，request对象是string，返回也为string|
||TorchPredict(TorchRequest)|向在线预测服务提交一个预测请求，request对象是TorchRequest类，返回为对应的TorchResponse|
||TFPredict(TFRequest)|向在线预测服务提交一个预测请求，request对象是TFRequest类，返回为对应的TFResponse|
|StringRequest|StringRequest{string("")}|TFRequest类构建方法，将string转换为StringRequest以调用Predict方法|
|TFRequest|TFRequest(signature_name)|TFRequest类构建方法，输入为要请求模型的signature_name|
||AddFeed$TYPE$(inputName string, shape []int64{}, tfDataType, content []$TYPE$)|请求Tensorflow的在线预测服务模型时，设置需要输入的Tensor，inputName表示输入Tensor的别名，tfDataType表示输入Tensor的DataType， shape表示输入Tensor的TensorShape，content表示输入Tensor的内容（一维数组展开表示）。DataType支持如下几种类型：easpredict.TfType_DT_FLOAT,easpredict.TfType_DT_DOUBLE,easpredict.TfType_DT_INT8,easpredict.TfType_DT_INT16,easpredict.TfType_DT_INT32,easpredict.TfType_DT_INT64,easpredict.TfType_DT_STRING,easpredict.TfType_DT_BOOL|
||AddFetch(outputName)|请求Tensorflow的在线预测服务模型时，设置需要输出的Tensor的别名，对于savedmodel模型该参数可选，若不设置，则输出所有的outputs，对于frozen model该参数必选|
|TFResponse|GetTensorShape(outputName)|获得别名为ouputname的输出Tensor的TensorShape|
||Get$TYPE$Val(outputName)|获取输出的tensor的数据向量，输出结果以一维数组的形式保存，可配套使用GetTensorShape()获取对应的tensor的shape，将其还原成所需的多维tensor, 其中$TYPE$可选Float, Double, Int, Int64, String, Bool|
|TorchRequest|TorchRequest()|TFRequest类构建方法|
||AddFeed(index, shape []int64{}, dataType, content)|请求PyTorch的在线预测服务模型时，设置需要输入的Tensor，index表示要输入的tensor的下标，dataType表示输入Tensor的DataType， shape表示输入Tensor的TensorShape，content表示输入Tensor的内容（一维数组展开表示）。DataType支持如下几种类型：easpredict.TorchType_DT_FLOAT, easpredict.TorchType_DT_DOUBLE, easpredict.TorchType_DT_INT32, easpredict.TorchType_DT_UINT8, easpredict.TorchType_DT_INT16, easpredict.TorchType_DT_INT8, easpredict.TorchType_DT_INT64, |
||AddFetch(outputIndex)|请求PyTorch的在线预测服务模型时，设置需要输出的Tensor的index，可选，若不设置，则输出所有的outputs|
|TorchResponse|GetTensorShape(outputIndex)|获得下标outputIndex的输出Tensor的TensorShape|
||Get$TYPE$Val(outputIndex)|获取输出的tensor的数据向量，输出结果以一维数组的形式保存，可配套使用GetTensorShape()获取对应的tensor的shape，将其还原成所需的多维tensor, $TYPE$可选Float, Double, Int, Int64|

# 程序示例


## String输入输出请求

```go
package main 

import (
	"fmt"
	"testing"
	"time"

	"eas-golang-sdk/easpredict"
)

func main() {
	client := easpredict.NewPredictClient("endpoind", "service_name")
	client.SetToken("token==")
	client.Init()
	// req := "test string"
	// for i := 0; i < 10; i++ {
	// 	resp, err := client.StringPredict(req)
	// 	if len(resp) == 0 || err != nil { // error handling
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println(resp)
	// }

	// general function
	req := StringRequest{"test string"}
	resp, err := client.Predict(req)

	if len(resp) == 0 || err != nil { // error handling
		fmt.Println(err)
		// ...
	}
	fmt.Println(resp.(*easpredict.StringResponse).str)

}
```

## TensorFlow输入输出请求

```go
package main 

import (
	"fmt"
	"testing"
	"time"

	"eas-golang-sdk/easpredict"
)

func main() {
	cli := easpredict.NewPredictClient("endpoint", "service_name")
	// cli.SetTimeout(1000)
	cli.Init()

	tfreq := easpredict.TFRequest{}
	tfreq.SetSignatureName("predict_images")
	tfreq.AddFeedFloat32("images", easpredict.TfType_DT_FLOAT, []int64{1, 784}, make([]float32, 784))
	// tfreq.AddFetch("scores")

	// st := time.Now()
	// for i := 0; i < 10; i++ {
	// 	resp, err := cli.TFPredict(tfreq)
	// 	if err != nil { // error handling
	// 		fmt.Println(err)
	// 		// ...
	// 	}
    //     fmt.Println(resp.GetTensorShape("scores"), resp.GetFloatVal("scores"))
	// }

	// fmt.Println("average response time : ", time.Since(st)/10)

	// general function:
	resp, err := cli.Predict(tfreq)

	if err != nil {
		fmt.Println("err", err)
		// ...
	}
	resp2, _ := resp.(*easpredict.TFResponse)
	fmt.Println(resp2.GetTensorShape("scores"), resp2.GetFloatVal("scores"))

}
```


## PyTorch输入输出程序示例
PyTorch用户可以使用TorchRequest与TorchResponse作为数据的输入输出格式，具体demo示例如下：

```go
package main 

import (
	"fmt"
	"testing"
	"time"

	"eas-golang-sdk/easpredict"
)

func main() {

	cli := easpredict.NewPredictClient("endpoint ", "service_name")

	// cli.SetTimeout(00)
	// cli.SetEndpointType("DIRECT")
	cli.Init()
	re := easpredict.TorchRequest{}
	re.AddFeedFloat32(0, easpredict.TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))
	re.AddFetch(0)

	// st := time.Now()
	// for i := 0; i < 10; i++ {
	// 	resp, err := cli.TorchPredict(re)
	// 	if err != nil { // error handling
	// 		fmt.Println(err)
	// 		// ...
	// 	}
	// 	fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
	// }
	// fmt.Println("average response time : ", time.Since(st)/10)

	// general function
	resp, err := cli.Predict(re)

	if err != nil {
		fmt.Println("err", err)
		// ...
	}
	resp2, _ := resp.(*easpredict.TorchResponse)
	fmt.Println(resp2.GetTensorShape(0), resp2.GetFloatVal(0))
}
```



**请用户注意：本sdk为beta版本，可供调试和试用。线上请谨慎试用！**

**请用户注意：本sdk为beta版本，可供调试和试用。线上请谨慎试用！**

**请用户注意：本sdk为beta版本，可供调试和试用。线上请谨慎试用！**
