package main

import (
	"fmt"

	"github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {
	client := eas.NewPredictClient("cn-beijing.pai-eas.aliyuncs.com", "test_dlrm")
	client.SetToken("tokenGeneratedFromService")
	client.Init()
	req := eas.TorchRequest{}
	
	length := 13312
	array := make([]int32, length)

	for i := range array {
		array[i] = 1
	}
	req.AddFeedMapFloat32("float_features", []int64{512,13}, make([]float32, 6656))
	req.AddFeedMapInt32("id_list_features.lengths", []int64{512,26}, array)
	req.AddFeedMapInt64("id_list_features.values", []int64{13312}, make([]int64, 13312))
	req.AddFetch(0)
	req.SetDebugLevel(903)
	for i := 0; i < 10; i++ {
		resp, err := client.TorchPredict(req)
		if err != nil {
			fmt.Printf("failed to predict: %v", err)
		} else {
			//fmt.Println(resp)
			fmt.Println(resp.GetTensorShapeMap("default"), resp.GetFloatValMap("default"))
		}
	}
}