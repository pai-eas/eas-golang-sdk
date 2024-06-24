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
	"sync"
)

var gzipWriterPool = &sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

var zlibWriterPool = &sync.Pool{
	New: func() interface{} {
		return zlib.NewWriter(nil)
	},
}

type CompressWriter interface {
	Reset(io.Writer)
	Write([]byte) (int, error)
	Close() error
}

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
	var w CompressWriter

	switch compressType {
	case CompressTypeGzip:
		// w = gzip.NewWriter(&b)
		w = gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(w)
		w.Reset(&b)
	case CompressTypeZlib:
		// w = zlib.NewWriter(&b)
		w = zlibWriterPool.Get().(*zlib.Writer)
		defer zlibWriterPool.Put(w)
		w.Reset(&b)
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
