package eas

import (
	// "eas-golang-sdk/easpredict"
	"fmt"
	"testing"
	"time"
)

const (
	EndpointName = ""
	PMMLName = ""
	PMMLToken = ""
	TensorflowName = ""
	TensorflowToken = ""
	TorchName = ""
	TorchToken = ""
)

func TestString(t *testing.T) {

	client := NewPredictClientWithConns(EndpointName, PMMLName, 10)
	client.SetToken(PMMLToken)
	client.Init()
	req := "[{}]"
	resp, err := client.StringPredict(req)
	if err != nil {
		t.Fatalf(err.Error())
	} else {
		fmt.Printf("%v\n", resp)
	}
}

func TestTF(t *testing.T) {
	cli := NewPredictClient(EndpointName, TensorflowName)
	cli.SetToken(TensorflowToken)
	cli.Init()

	tfreq := TFRequest{}
	tfreq.SetSignatureName("predict_images")
	tfreq.AddFeedFloat32("images", TfType_DT_FLOAT, []int64{1, 784}, make([]float32, 784))

	st := time.Now()
	for i := 0; i < 10; i++ {
		resp, err := cli.TFPredict(tfreq)
		if err != nil {
			t.Fatalf("failed to query tf model: %v", err)
		}
		fmt.Printf("%v\n", resp)
	}

	fmt.Println("average response time : ", time.Since(st)/10)
}

// TestTorch tests pytorch request and response unit test
func TestTorch(t *testing.T) {

	cli := NewPredictClient(EndpointName, TorchName)
	cli.SetTimeout(80)
	cli.SetRetryCount(5)
	cli.SetToken(TorchToken)
	cli.Init()
	re := TorchRequest{}
	re.AddFeedFloat32(0, TorchType_DT_FLOAT, []int64{1, 3, 224, 224}, make([]float32, 150528))
	re.AddFetch(0)
	st := time.Now()
	for i := 0; i < 10; i++ {
		resp, err := cli.TorchPredict(re)
		if err != nil {
			t.Fatalf("failed to query torch model: %v", err)
		}
		fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
	}
	fmt.Println("average response time : ", time.Since(st)/10)
}
