// package easpredict

package main

import (
	"eas-golang-sdk/easpredict"
	"fmt"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	client := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com", "randsleep_multi_instance")
	client.SetToken("MmNiYzNlYTU4NDU3YmI2NzgyMjhiZTI3YmExZjA0YTYyYzg5ZmI0MQ==")
	client.Init()
	req := "test string"
	for i := 0; i < 100; i++ {
		go fmt.Println(client.StringPredict(req))
		fmt.Print(i)
	}
}

// TestTorch tests pytorch request and response unit test
func TestTorch(t *testing.T) {

	// cli := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch-wl-gosdktest")
	cli := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch_gpu_wl")

	// cli.SetTimeout(800)
	// cli.SetEndpointType("DIRECT")
	cli.Init()
	re := easpredict.TorchRequest{}
	re.AddFeedFloat32(0, easpredict.TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))
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

func TestTF(t *testing.T) {
	// cli := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com/", "tf_gosdk_test")
	// cli.SetToken("BTX3ZLQ5lzlkMzYnnMo0PzV5Yzc0YzB9M3ZhNzM5B2iwUjU4Y2MwXA==")

	cli := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com", "warm_up_test")
	cli.SetToken("NDBiZGEyZDBjMzVlNjM3NDMxYTc1NGNmZGFlMzMzMjBmMmY2MzE3ZQ==")
	// cli.SetTimeout(1000)
	cli.Init()

	tfreq := easpredict.TFRequest{}
	// tfreq.SetSignatureName("predict_images")
	// tfreq.AddFeedFloat32("images", easpredict.TfType_DT_FLOAT, []int64{1, 784}, make([]float32, 784))
	// tfreq.AddFetch("scores")
	// fmt.Println(tfreq)

	// ls := list.New()
	// ls.PushBack("abcdef")
	tfreq.SetSignatureName("serving_default")
	tfreq.AddFeedString("input_holder", easpredict.TfType_DT_STRING, []int64{1}, [][]byte{[]byte("abcdef")})
	// th := list.New()
	// th.PushBack(0.9)
	tfreq.AddFeedFloat32("threshold", easpredict.TfType_DT_FLOAT, []int64{}, []float32{0.9})
	// model_name := list.New()
	// model_name.PushBack("PACKAGE_640")
	tfreq.AddFeedString("model_id", easpredict.TfType_DT_STRING, []int64{}, [][]byte{[]byte("PACKAGE_640")})
	// tfreq.AddFetch("sorted_probs")
	// tfreq.AddFetch("sorted_labels")

	st := time.Now()
	for i := 0; i < 100; i++ {
		resp := cli.TFPredict(tfreq)
		// fmt.Println(resp.GetTensorShape("sorted_probs"), resp.GetFloatVal("sorted_probs"))
		fmt.Println(resp.GetFloatVal("sorted_probs"))
	}

	fmt.Println("average response time : ", time.Since(st)/10)
}

// // TestTorch tests pytorch request and response
// func TestTorchVIP(t *testing.T) {

// 	// cli := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch-wl-gosdktest")
// 	cli := easpredict.NewPredictClient("eas-shanghai.alibaba-inc.com", "pytorch_gpu_wl")
// 	cli.SetEndpointType("VIPSERVER")
// 	// cli.SetEndpointType("DIRECT")
// 	cli.Init()
// 	re := easpredict.TorchRequest{}

// 	re.AddFeedFloat32(0, easpredict.TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))

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
