package eas

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

func md5sum(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func hmacSha256(data string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func compress(data []byte, compressType string) ([]byte, error) {
	var b bytes.Buffer
	var w io.WriteCloser

	switch compressType {
	case CompressTypeGzip:
		w = gzip.NewWriter(&b)
	case CompressTypeZlib:
		w = zlib.NewWriter(&b)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressType)
	}

	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func noPanic() {
	_ = recover()
}
