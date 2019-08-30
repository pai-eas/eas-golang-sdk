package easprediction

import "./tf_predict_protos"

// TfRequest class for tensorflow data and requests
type TfRequest struct {
	RequestData tf_predict_protos.PredictRequest
}

// SetSignatureName set signature name for TensorFlow request
func (tr *TfRequest) SetSignatureName(sigName string) {
	tr.RequestData.SignatureName = sigName
}

// AddFeedFloat32 function adds float values input data for TfRequest
func (tr *TfRequest) AddFeedFloat32(inputName string, tfDataType tf_predict_protos.ArrayDataType, shape []int64, content []float32) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: tfDataType,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		FloatVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedFloat64 function adds double values input data for TfRequest
func (tr *TfRequest) AddFeedFloat64(inputName string, tfDataType tf_predict_protos.ArrayDataType, shape []int64, content []float64) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: tfDataType,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		DoubleVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedInt32 function adds int values input data for TfRequest
func (tr *TfRequest) AddFeedInt32(inputName string, tfDataType tf_predict_protos.ArrayDataType, shape []int64, content []int32) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: tfDataType,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		IntVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedInt64 function adds int64 values input data for TfRequest
func (tr *TfRequest) AddFeedInt64(inputName string, tfDataType tf_predict_protos.ArrayDataType, shape []int64, content []int64) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: tfDataType,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		Int64Val: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedBool function adds boolean values input data for TfRequest
func (tr *TfRequest) AddFeedBool(inputName string, tfDataType tf_predict_protos.ArrayDataType, shape []int64, content []bool) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: tfDataType,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		BoolVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedString function adds string values input data for TfRequest
func (tr *TfRequest) AddFeedString(inputName string, tfDataType tf_predict_protos.ArrayDataType, shape []int64, content [][]byte) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: tfDataType,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		StringVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFetch adds output filter (outname) for TensorFlow request
func (tr *TfRequest) AddFetch(outName string) {
	tr.RequestData.OutputFilter = append(tr.RequestData.OutputFilter, outName)
}

// TfResponse class for Pytf predicted results
type TfResponse struct {
	Response tf_predict_protos.PredictResponse
}

// GetTensorShape returns []int64 slice as shape of tensor outindexed
func (tresp *TfResponse) GetTensorShape(outputName string) []int64 {
	// return tresp.PredictResponse.Outputs[outputName].ArrayShape.Dim
	return tresp.Response.Outputs[outputName].ArrayShape.Dim
}

// GetFloatVal returns []float32 slice as output data
func (tresp *TfResponse) GetFloatVal(outputName string) []float32 {
	return tresp.Response.Outputs[outputName].GetFloatVal()
}

// GetDoubleVal returns []float64 slice as output data
func (tresp *TfResponse) GetDoubleVal(outputName string) []float64 {
	return tresp.Response.Outputs[outputName].GetDoubleVal()
}

// GetIntVal returns []int32 slice as output data
func (tresp *TfResponse) GetIntVal(outputName string) []int32 {
	return tresp.Response.Outputs[outputName].GetIntVal()
}

// GetInt64Val returns []int64 slice as output data
func (tresp *TfResponse) GetInt64Val(outputName string) []int64 {
	return tresp.Response.Outputs[outputName].GetInt64Val()
}

// GetBoolVal returns []bool slice as output data
func (tresp *TfResponse) GetBoolVal(outputName string) []bool {
	return tresp.Response.Outputs[outputName].GetBoolVal()
}

// GetStringVal returns []string slice as output data
func (tresp *TfResponse) GetStringVal(outputName string) [][]byte {
	return tresp.Response.Outputs[outputName].GetStringVal()
}
