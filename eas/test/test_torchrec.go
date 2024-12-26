package main

import (
	"fmt"

	"github.com/pai-eas/eas-golang-sdk/eas"
)

func main() {

	client := eas.NewPredictClient("cn-beijing.pai-eas.aliyuncs.com", "test_torch_rec_multi_tower_din_gpu")
	client.SetToken("tokenGeneratedFromService")
	client.Init()
	req := eas.TorchRecRequest{}

	req.AddItemId("7033")
	req.AddUserFeature("user_id",33981,"int")

	length := 4
	array := make([]float64, length)
	array[0] = 0.24689289764507472
	array[1] = 0.005758482924454689
	array[2] = 0.6765301324940026
	array[3] = 0.18137273055602343
	req.AddUserFeature("raw_3",array,"list<double>")

	myMap := make(map[string]int32)
	myMap["866"] = 4143
	myMap["1627"] = 2451
	req.AddUserFeature("map_2",myMap,"map<string,int>")

	rows := 3
	cols := 4
	array2 := make([][]float32, rows)
	for i := range array2 {
		array2[i] = make([]float32, cols)
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			array2[i][j] = 1.0
		}
	}
	req.AddUserFeature("click",array2,"list<list<float>>")

	req.AddContextFeature("id_2",array,"list<double>")
	req.AddContextFeature("id_2",array,"list<double>")

	req.AddContextFeature("id_3","a","string")
	req.AddContextFeature("id_3","b","string")

	req.AddItemFeature("id_4","a","string")
	req.AddItemFeature("id_4","b","string")

	req.SetDebugLevel(903)
	fmt.Println(req)
	for i := 0; i < 10; i++ {
		resp, err := client.TorchRecPredict(req)
		if err != nil {
			fmt.Printf("failed to predict: %v", err)
		} else {
			fmt.Println(resp.GetTensorShapeMap("logits"), resp.GetFloatValMap("logits"))
		}
	}
}
