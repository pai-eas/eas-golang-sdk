package types

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func Test_FB_Codec_Object(t *testing.T) {
	codec := &fbDataFrameCodec{}

	df := &DataFrame{}
	df.Tags = map[string]string{
		"foo": "bar",
	}
	df.Index = 10000
	df.Data = []byte("hello world")
	sb := &bytes.Buffer{}
	err := codec.Encode(*df, sb)
	if err != nil {
		t.Fatal(err)
	}
	df2 := &DataFrame{}
	err = codec.Decode(sb.Bytes(), df2)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(df, df2) == false {
		t.Fatal("not equal")
	}
}

func Benchmark_PB_Codec_Object(b *testing.B) {
	codec := &pbDataFrameCodec{}
	df := &DataFrame{}
	sb := &bytes.Buffer{}
	df.Tags = map[string]string{"foo": "bar"}
	df.Index = 10000
	df.Data = []byte("hello world")
	df.Message = "12345"
	err := codec.Encode(*df, sb)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		err = codec.Decode(sb.Bytes(), &DataFrame{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_FB_Codec_Object(b *testing.B) {
	codec := &fbDataFrameCodec{}
	df := &DataFrame{}
	sb := &bytes.Buffer{}
	df.Tags = map[string]string{"foo": "bar"}
	df.Index = 10000
	df.Data = []byte("hello world")
	df.Message = "12345"
	err := codec.Encode(*df, sb)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		err = codec.Decode(sb.Bytes(), &DataFrame{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Test_FB_Codec_List(t *testing.T) {
	codec := &fbDataFrameCodec{}
	var list []DataFrame
	for i := 0; i < 3; i++ {
		df := &DataFrame{}
		df.Tags = map[string]string{
			"foo": "bar",
		}
		df.Message = "12345"
		df.Index = Index(10000 + uint64(i))
		df.Data = []byte("hello world" + fmt.Sprint(i))
		list = append(list, *df)
	}

	sb := &bytes.Buffer{}
	err := codec.EncodeList(list, sb)
	if err != nil {
		t.Fatal(err)
	}
	list2, err := codec.DecodeList(sb.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Index < list[j].Index
	})
	sort.Slice(list2, func(i, j int) bool {
		return list2[i].Index < list2[j].Index
	})

	if reflect.DeepEqual(list, list2) == false {
		t.Fatalf("not equal: %v vs %v", list, list2)
	}
}

func Test_FB_Codec_Attr(t *testing.T) {
	codec := &fbAttributesCodec{}
	attr := Attributes{
		"foo": "bar",
	}

	sb := &bytes.Buffer{}
	err := codec.Encode(attr, sb)
	if err != nil {
		t.Fatal(err)
	}

	attr2 := Attributes{}
	err = codec.Decode(sb.Bytes(), &attr2)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(attr, attr2) == false {
		t.Fatalf("not equal: %v vs %v", attr, attr2)
	}
}

func Test_JSON_Codec_Attr(t *testing.T) {
	codec := &jsonAttributesCodec{}
	attr := Attributes{
		"foo": "bar",
	}
	sb := &bytes.Buffer{}
	err := codec.Encode(attr, sb)
	if err != nil {
		t.Fatal(err)
	}

	attr2 := Attributes{}
	err = codec.Decode(sb.Bytes(), &attr2)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(attr, attr2) == false {
		t.Fatalf("not equal: %v vs %v", attr, attr2)
	}
}
