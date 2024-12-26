package eas

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pai-eas/eas-golang-sdk/eas/types/torch_predict_protos"
)

// TorchRecRequest class for PyTorch data and requests
type TorchRecRequest struct {
    RequestData torch_predict_protos.PBRequest
}

// Set Debug level for torchrecrequest
func (tr *TorchRecRequest) SetDebugLevel(debug_level int32) {
    tr.RequestData.DebugLevel = debug_level
}

// Set FaissNeighNum for torchrecrequest
func (tr *TorchRecRequest) SetFaissNeighNum(k int32) {
    tr.RequestData.FaissNeighNum = k
}

func (tr *TorchRecRequest) AddFeat(value interface{}, dtype string) (*torch_predict_protos.PBFeature, error) {
    dtype = strings.ToUpper(dtype)
    feat := &torch_predict_protos.PBFeature{}

    if value == nil || isEmpty(value) {
        feat.Value = &torch_predict_protos.PBFeature_StringFeature{StringFeature: ""}
    } else if dtype == "STRING" {
        feat.Value = &torch_predict_protos.PBFeature_StringFeature{StringFeature: value.(string)} 
    } else if dtype == "FLOAT" {
        feat.Value = &torch_predict_protos.PBFeature_FloatFeature{FloatFeature: float32(value.(float64))} 
    } else if dtype == "DOUBLE" {
        feat.Value = &torch_predict_protos.PBFeature_DoubleFeature{DoubleFeature: value.(float64)} 
    } else if dtype == "BIGINT" || dtype == "INT64" {
        feat.Value = &torch_predict_protos.PBFeature_LongFeature{LongFeature: value.(int64)} 
    } else if dtype == "INT" {
        feat.Value = &torch_predict_protos.PBFeature_IntFeature{IntFeature: int32(value.(int))}
    } else if isListType(dtype) {
        return tr.AddToListField(feat,value, dtype)
    } else if isMapType(dtype) {
        return tr.AddToMapField(feat, value, dtype)
    } else if dtype == "ARRAY<ARRAY<FLOAT>>" || dtype == "LIST<LIST<FLOAT>>" {
        if lists, ok := value.([][]float32); ok {
            flists := &torch_predict_protos.FloatLists{}
            for _, sublist := range lists {
                list := &torch_predict_protos.FloatList{}
                for _, v := range sublist {
                    list.Features = append(list.Features, v)
                }
                flists.Lists = append(flists.Lists, list) 
            }
            feat.Value = &torch_predict_protos.PBFeature_FloatLists{FloatLists: flists}
        } else {
            return nil, errors.New("Expected value to be a list of lists for ARRAY<ARRAY<FLOAT>>/LIST<LIST<FLOAT>> dtype")
        }
    } else {
        return nil, fmt.Errorf("unsupported dtype: %s", dtype)
    }
    return feat, nil
}

func isMap(value interface{}) bool {
    val := reflect.ValueOf(value)
    return val.Kind() == reflect.Map 
}

func isList(value interface{}) bool {
    val := reflect.ValueOf(value)
    return val.Kind() == reflect.Slice
}

func isEmpty(value interface{}) bool {
    val := reflect.ValueOf(value)
    if !val.IsValid() {
        return true
    }

    if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
        return val.IsNil()
    }
    return (isMap(value) && val.Len() == 0) || (isList(value) && val.Len() == 0)
}

func isListType(dtype string) bool {
    validTypes := []string{
        "LIST<FLOAT>",
        "LIST<STRING>",
        "LIST<DOUBLE>",
        "LIST<INT>",
        "LIST<INT64>",
        "LIST<BIGINT>",
        "ARRAY<FLOAT>",
        "ARRAY<STRING>",
        "ARRAY<DOUBLE>",
        "ARRAY<INT>",
        "ARRAY<INT64>",
        "ARRAY<BIGINT>",
    }
    for _, validType := range validTypes {
        if validType == dtype {
            return true
        }
    }
    return false
}

func isMapType(dtype string) bool {
    validTypes := []string{
        "MAP<INT,INT>","MAP<INT,INT64>","MAP<INT,BIGINT>","MAP<INT,STRING>","MAP<INT,FLOAT>","MAP<INT,DOUBLE>" ,
        "MAP<INT64,INT>","MAP<INT64,INT64>","MAP<INT64,BIGINT>","MAP<INT64,STRING>","MAP<INT64,FLOAT>","MAP<INT64,DOUBLE>" ,
        "MAP<BIGINT,INT>","MAP<BIGINT,INT64>","MAP<BIGINT,BIGINT>","MAP<BIGINT,STRING>","MAP<BIGINT,FLOAT>","MAP<BIGINT,DOUBLE>",
        "MAP<STRING,INT>","MAP<STRING,INT64>","MAP<STRING,BIGINT>","MAP<STRING,STRING>","MAP<STRING,FLOAT>","MAP<STRING,DOUBLE>",
    }

    for _, validType := range validTypes {
        if validType == dtype {
            return true
        }
    }

    return false
}

func (tr *TorchRecRequest) AddToListField(feat *torch_predict_protos.PBFeature, value interface{}, dtype string) (*torch_predict_protos.PBFeature, error) {
    switch dtype {
    case "LIST<STRING>", "ARRAY<STRING>":
        list := &torch_predict_protos.StringList{}
        if vals, ok := value.([]string); ok {
            for _, v := range vals {
                list.Features = append(list.Features, v) 
            }
            feat.Value = &torch_predict_protos.PBFeature_StringList{StringList: list}
        } else {
            return nil, fmt.Errorf("value must be of type []string")
        }

    case "LIST<FLOAT>", "ARRAY<FLOAT>":
        list := &torch_predict_protos.FloatList{}
        if vals, ok := value.([]float32); ok {
            for _, v := range vals {
                list.Features = append(list.Features, v)
            }
            feat.Value = &torch_predict_protos.PBFeature_FloatList{FloatList: list}
        } else {
            return nil, fmt.Errorf("value must be of type []float32")
        }

    case "LIST<DOUBLE>", "ARRAY<DOUBLE>":
        list := &torch_predict_protos.DoubleList{}
        if vals, ok := value.([]float64); ok {
            for _, v := range vals {
                list.Features = append(list.Features, v)
            }
            feat.Value = &torch_predict_protos.PBFeature_DoubleList{DoubleList: list}
        } else {
            return nil, fmt.Errorf("value must be of type []float64")
        }

    case "LIST<INT64>", "ARRAY<INT64>", "LIST<BIGINT>", "ARRAY<BIGINT>":
        list := &torch_predict_protos.LongList{}
        if vals, ok := value.([]int64); ok {
            for _, v := range vals {
                list.Features = append(list.Features, v)
            }
            feat.Value = &torch_predict_protos.PBFeature_LongList{LongList: list}
        } else {
            return nil, fmt.Errorf("value must be of type []int64")
        }

    case "LIST<INT>", "ARRAY<INT>":
        list := &torch_predict_protos.IntList{}
        if vals, ok := value.([]int32); ok {
            for _, v := range vals {
                list.Features = append(list.Features, v)
            }
            feat.Value = &torch_predict_protos.PBFeature_IntList{IntList: list}
        } else {
            return nil, fmt.Errorf("value must be of type []int32")
        }

    default:
        return nil, fmt.Errorf("unsupported dtype: %s", dtype)
    }

    return feat, nil
}

func (tr *TorchRecRequest) AddToMapField(feat *torch_predict_protos.PBFeature, value interface{}, dtype string) (*torch_predict_protos.PBFeature, error) {
    switch dtype {
	//int -> others
    case "MAP<INT,INT>":
        mymap := &torch_predict_protos.IntIntMap{MapField: make(map[int32]int32)}
        if vals, ok := value.(map[int32]int32); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_IntIntMap{IntIntMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int32]int32 for MAP type, got %T", value)
        }

    case "MAP<INT,INT64>", "MAP<INT,BIGINT>":
        mymap := &torch_predict_protos.IntLongMap{MapField: make(map[int32]int64)}
        if vals, ok := value.(map[int32]int64); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_IntLongMap{IntLongMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int32]int64 for MAP type, got %T", value)
        }

    case "MAP<INT,STRING>":
        mymap := &torch_predict_protos.IntStringMap{MapField: make(map[int32]string)}
        if vals, ok := value.(map[int32]string); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_IntStringMap{IntStringMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int32]string for MAP type, got %T", value)
        }
    case "MAP<INT,FLOAT>":
        mymap := &torch_predict_protos.IntFloatMap{MapField: make(map[int32]float32)}
        if vals, ok := value.(map[int32]float32); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_IntFloatMap{IntFloatMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int32]float32 for MAP type, got %T", value)
        }

    case "MAP<INT,DOUBLE>":
        mymap := &torch_predict_protos.IntDoubleMap{MapField: make(map[int32]float64)}
        if vals, ok := value.(map[int32]float64); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_IntDoubleMap{IntDoubleMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int32]float64 for MAP type, got %T", value)
        }
	
	//string -> others
	case "MAP<STRING,INT>":
        mymap := &torch_predict_protos.StringIntMap{MapField: make(map[string]int32)}
        if vals, ok := value.(map[string]int32); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_StringIntMap{StringIntMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[string]int32 for MAP type, got %T", value)
        }

    case "MAP<STRING,INT64>", "MAP<STRING,BIGINT>":
        mymap := &torch_predict_protos.StringLongMap{MapField: make(map[string]int64)}
        if vals, ok := value.(map[string]int64); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_StringLongMap{StringLongMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[string]int64 for MAP type, got %T", value)
        }

    case "MAP<STRING,STRING>":
        mymap := &torch_predict_protos.StringStringMap{MapField: make(map[string]string)}
        if vals, ok := value.(map[string]string); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_StringStringMap{StringStringMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[string]string for MAP type, got %T", value)
        }
    case "MAP<STRING,FLOAT>":
        mymap := &torch_predict_protos.StringFloatMap{MapField: make(map[string]float32)}
        if vals, ok := value.(map[string]float32); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_StringFloatMap{StringFloatMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[string]float32 for MAP type, got %T", value)
        }

    case "MAP<STRING,DOUBLE>":
        mymap := &torch_predict_protos.StringDoubleMap{MapField: make(map[string]float64)}
        if vals, ok := value.(map[string]float64); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_StringDoubleMap{StringDoubleMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[string]float64 for MAP type, got %T", value)
        }

	//int64->others
	case "MAP<INT64,INT>","MAP<BIGINT,INT>":
        mymap := &torch_predict_protos.LongIntMap{MapField: make(map[int64]int32)}
        if vals, ok := value.(map[int64]int32); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_LongIntMap{LongIntMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int64]int32 for MAP type, got %T", value)
        }

    case "MAP<INT64,INT64>","MAP<BIGINT,BIGINT>","MAP<INT64,BIGINT>","MAP<BIGINT,INT64>":
        mymap := &torch_predict_protos.LongLongMap{MapField: make(map[int64]int64)}
        if vals, ok := value.(map[int64]int64); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_LongLongMap{LongLongMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int64]int64 for MAP type, got %T", value)
        }

    case "MAP<INT64,STRING>","MAP<BIGINT,STRING>":
        mymap := &torch_predict_protos.LongStringMap{MapField: make(map[int64]string)}
        if vals, ok := value.(map[int64]string); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_LongStringMap{LongStringMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int64]string for MAP type, got %T", value)
        }
    case "MAP<INT64,FLOAT>","MAP<BIGINT,FLOAT>":
        mymap := &torch_predict_protos.LongFloatMap{MapField: make(map[int64]float32)}
        if vals, ok := value.(map[int64]float32); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_LongFloatMap{LongFloatMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int64]float32 for MAP type, got %T", value)
        }

    case "MAP<INT64,DOUBLE>","MAP<BIGINT,DOUBLE>":
        mymap := &torch_predict_protos.LongDoubleMap{MapField: make(map[int64]float64)}
        if vals, ok := value.(map[int64]float64); ok {
            for k, v := range vals {
                mymap.MapField[k] = v
            }
            feat.Value = &torch_predict_protos.PBFeature_LongDoubleMap{LongDoubleMap: mymap}
        } else {
            return nil, fmt.Errorf("expected map[int64]float64 for MAP type, got %T", value)
        }

    default:
        return nil, fmt.Errorf("unsupported dtype: %s", dtype)
    }

    return feat, nil
}

// add user features for torchrecrequest
func (tr *TorchRecRequest) AddUserFeature(key string,value interface{}, dtype string)  {
    feat, err := tr.AddFeat(value, dtype)
    if err != nil{
        fmt.Println("failed to add user feature, key:",key," the err is:",err)
        return 
    }
    
    if tr.RequestData.UserFeatures == nil {
        tr.RequestData.UserFeatures = make(map[string]*torch_predict_protos.PBFeature) 
    }
    tr.RequestData.UserFeatures[key] = feat
    
}

// add context features for torchrecrequest
func (tr *TorchRecRequest) AddContextFeature(key string,value interface{}, dtype string) {
    feat, err := tr.AddFeat(value, dtype)
    if err != nil{
        fmt.Println("failed to add context feature, key:",key," the err is:",err)
        return 
    }
    if tr.RequestData.ContextFeatures == nil {
        tr.RequestData.ContextFeatures = make(map[string]*torch_predict_protos.ContextFeatures)
    }

    if tr.RequestData.ContextFeatures[key] == nil {
        tr.RequestData.ContextFeatures[key] = &torch_predict_protos.ContextFeatures{}
    }

    tr.RequestData.ContextFeatures[key].Features = append(tr.RequestData.ContextFeatures[key].Features, feat)
}

// add item features for torchrecrequest
func (tr *TorchRecRequest) AddItemFeature(key string,value interface{}, dtype string) {
    feat, err := tr.AddFeat(value, dtype)
    if err != nil{
        fmt.Println("failed to add context feature, key:",key," the err is:",err)
        return 
    }
    if tr.RequestData.ItemFeatures == nil {
        tr.RequestData.ItemFeatures = make(map[string]*torch_predict_protos.ContextFeatures)
    }

    if tr.RequestData.ItemFeatures[key] == nil {
        tr.RequestData.ItemFeatures[key] = &torch_predict_protos.ContextFeatures{}
    }

    tr.RequestData.ItemFeatures[key].Features = append(tr.RequestData.ItemFeatures[key].Features, feat)
}

// add item ids for torchrecrequest
func (tr *TorchRecRequest) AddItemId(itemId string) {
    tr.RequestData.ItemIds = append(tr.RequestData.ItemIds, itemId)
}



// ToString for interface
func (tr TorchRecRequest) ToString() (string, error) {
    reqData, err := proto.Marshal(&tr.RequestData)
    if err != nil {
        return "", NewPredictError(-1, "", err.Error())
    }
    return string(reqData), nil
}

// TorchResponse class for PyTorch predicted results
type TorchRecResponse struct {
    Response torch_predict_protos.PBResponse
}

// GetTensorShape returns []int64 slice as shape of tensor outindexed
func (resp *TorchRecResponse) GetTensorShapeMap(outIndex string) []int64 {
    return resp.Response.MapOutputs[outIndex].ArrayShape.Dim
}

// GetFloatValMap returns []float32 slice as output data
func (resp *TorchRecResponse) GetFloatValMap(outIndex string) []float32 {
    return resp.Response.MapOutputs[outIndex].GetFloatVal()
}

// GetDoubleValMap returns []float64 slice as output data
func (resp *TorchRecResponse) GetDoubleValMap(outIndex string) []float64 {
	return resp.Response.MapOutputs[outIndex].GetDoubleVal()
}

// GetIntValMap returns []int32 slice as output data
func (resp *TorchRecResponse) GetIntValMap(outIndex string) []int32 {
    return resp.Response.MapOutputs[outIndex].GetIntVal()
}

// GetInt64ValMap returns []int64 slice as output data
func (resp *TorchRecResponse) GetInt64ValMap(outIndex string) []int64 {
    return resp.Response.MapOutputs[outIndex].GetInt64Val()
}

// Unmarshal for interface
func (resp *TorchRecResponse) unmarshal(body []byte) error {
    bd := &torch_predict_protos.PBResponse{}
    err := proto.Unmarshal(body, bd)
    if err != nil {
        return err
    }
    resp.Response = *bd
    return nil
}
