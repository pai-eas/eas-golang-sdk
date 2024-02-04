package eas

import (
	"bytes"
	"io"
	"net/http"
)

type RawRequest struct {
	http.Request
}

func (r RawRequest) ToString() (string, error) {
	if r.Request.Body == nil {
		return "", nil
	}
	defer r.Request.Body.Close()

	rawBytes, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return "", err
	}

	r.Request.Body = io.NopCloser(bytes.NewReader(rawBytes))

	return string(rawBytes), nil
}

type RawResponse struct {
	data []byte
}

func (r RawResponse) unmarshal(ret []byte) error {
	// there is no need to unmarshal, just skip
	return nil
}

func (r RawResponse) ToBytes() []byte {
	return r.data
}
