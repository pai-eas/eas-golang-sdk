package easpredict

import (
	// "eas-golang-sdk/easpredict"
	"fmt"
	"testing"
	"time"
)

func TestString(t *testing.T) {

	client := NewPredictClientWithConns("proxyed.shanghai.eas.vipserver", "", 10)
	client.SetToken("1111111111")
	// client.SetEndpointType("DIRECT")
	client.SetEndpointType("VIPSERVER")
	client.Init()
	req := "random string"
	for i := 0; i < 1; i++ {
		go func(i int) {
			resp, err := client.StringPredict(req)
			fmt.Print(i)
			fmt.Println(resp, "er", err)
		}(i)
	}
	time.Sleep(time.Duration(10) * time.Second)
}

// TestTorch tests pytorch request and response unit test
func TestTorch(t *testing.T) {

	cli := NewPredictClient("eas-shanghai.alibaba-inc.com/api/predict/test624_cpu", "")

	cli.SetTimeout(80)
	// cli.SetEndpointType("DIRECT")
	cli.SetToken("ZTk1OThjM2YyNzkwNDZiZTI5YTNjMmY5NzYxNmQxZTQ3MjUyYzA4Zg==")
	cli.Init()
	re := TorchRequest{}
	re.AddFeedFloat32(0, TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))
	re.AddFetch(0)
	st := time.Now()
	for i := 0; i < 10; i++ {
		resp, err := cli.TorchPredict(re)
		if err != nil {
			fmt.Println("err", err)
		}
		// if resp.GetTensorShape(0) != nil {
		// 	t.Log("predict success: ", resp.GetTensorShape(0))
		// }
		fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
	}
	fmt.Println("average response time : ", time.Since(st)/10)
}

func TestTF(t *testing.T) {
	cli := NewPredictClient("endpoint", "service_name")
	cli.SetToken("token==")
	// cli.SetTimeout(1000)
	cli.Init()

	tfreq := TFRequest{}

	tfreq.SetSignatureName("serving_default")
	tfreq.AddFeedString("input_holder", TfType_DT_STRING, []int64{1}, [][]byte{[]byte("abcdef")})
	tfreq.AddFeedFloat32("threshold", TfType_DT_FLOAT, []int64{}, []float32{0.9})
	tfreq.AddFeedString("model_id", TfType_DT_STRING, []int64{}, [][]byte{[]byte("PACKAGE_640")})
	// tfreq.AddFetch("sorted_probs")
	// tfreq.AddFetch("sorted_labels")

	st := time.Now()
	for i := 0; i < 1; i++ {
		resp, err := cli.TFPredict(tfreq)
		// fmt.Println(resp.GetTensorShape("sorted_probs"), resp.GetFloatVal("sorted_probs"))
		fmt.Println(resp.GetFloatVal("sorted_probs"), err)
	}

	fmt.Println("average response time : ", time.Since(st)/10)
}

// func main() {
// 	// testTorch()
// }
