package main

import (
	"fmt"
	"testing"
	"time"

	"./easprediction"
)

// TestTorch tests pytorch request and response unit test
func TestTorch(t *testing.T) {

	// cli := easprediction.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch-wl-gosdktest")
	cli := easprediction.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch_gpu_wl")

	// cli.SetTimeout(800)
	// cli.SetEndpointType("DIRECT")
	cli.Init()
	re := easprediction.TorchRequest{}
	re.AddFeedFloat32(0, easprediction.TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))
	re.AddFetch(0)
	st := time.Now()
	for i := 0; i < 10; i++ {
		resp := cli.TorchPredict(re)
		if resp.GetTensorShape(0) != nil {
			t.Log("predict success: ", resp.GetTensorShape(0))
		}
		// fmt.Println(resp, err)
		// fmt.Println(resp.GetFloatVal(0))
		fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
	}
	fmt.Println("average response time : ", time.Since(st)/10)
}

func TestTf(t *testing.T) {
	cli := easprediction.NewPredictClient("eas-shanghai.alibaba-inc.com/", "tf_gosdk_test")
	// cli.SetToken("BTX3ZLQ5lzlkMzYnnMo0PzV5Yzc0YzB9M3ZhNzM5B2iwUjU4Y2MwXA==")
	// cli.SetTimeout(1000)
	cli.Init()

	tfreq := easprediction.TfRequest{}
	tfreq.SetSignatureName("predict_images")
	tfreq.AddFeedFloat32("images", easprediction.TfType_DT_FLOAT, []int64{1, 784}, make([]float32, 784))
	// tfreq.AddFetch("scores")
	// fmt.Println(tfreq)

	st := time.Now()
	for i := 0; i < 10; i++ {
		resp := cli.TfPredict(tfreq)
		fmt.Println(resp.GetTensorShape("scores"), resp.GetFloatVal("scores"))
	}

	fmt.Println("average response time : ", time.Since(st)/10)
}

// // TestTorch tests pytorch request and response
// func TestTorchVIP(t *testing.T) {

// 	// cli := easprediction.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch-wl-gosdktest")
// 	cli := easprediction.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch_gpu_wl")
// 	cli.SetEndpointType("VIPSERVER")
// 	// cli.SetEndpointType("DIRECT")
// 	cli.Init()
// 	re := easprediction.TorchRequest{}

// 	re.AddFeedFloat32(0, easprediction.TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))

// 	for i := 0; i < 10; i++ {

// 		resp := cli.TorchPredict(re)
// 		if resp.GetTensorShape(0) != nil {
// 			t.Log("get tensor 0: ", resp.GetTensorShape(0))
// 		}
// 		// fmt.Println(resp, err)
// 		// fmt.Println(resp.GetFloatVal(0))
// 		fmt.Println(resp.GetTensorShape(0))
// 	}
// }

func main() {
	// testTorch()
}
