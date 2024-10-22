package main

import (
	"fmt"

	"github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {

	client := eas.NewPredictClient("cn-beijing.pai-eas.aliyuncs.com", "test_torch_rec_din_debug")
	client.SetToken("tokenGeneratedFromService")
	client.Init()
	req := eas.TorchRequest{}
	req.AddFeedInt64(0, []int64{300,3}, make([]int64, 900))
	req.AddFeedFloat32(1, []int64{300,10,768}, make([]float32, 2304000))
	req.AddFeedFloat32(2, []int64{300,768}, make([]float32, 230400))
	req.AddFeedInt64(3, []int64{300}, make([]int64, 300))
	req.AddFetch(0)
	req.SetDebugLevel(903)
	for i := 0; i < 10; i++ {
		resp, err := client.TorchPredict(req)
		if err != nil {
			fmt.Printf("failed to predict: %v", err)
		} else {
			fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
		}
	}
}