package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Tags map[string]string

func (t Tags) String() string {
	var allKeys []string
	for key := range t {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)
	var sb strings.Builder
	var count int
	sb.WriteString("tags")
	sb.WriteByte('[')
	for _, key := range allKeys {
		if count > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(key + "=" + t[key])
		count++
	}
	sb.WriteByte(']')
	return sb.String()
}

func (t Tags) Validate() error {
	if len(t) == 0 {
		return nil
	}
	for key := range t {
		if strings.HasPrefix(key, "_") {
			return fmt.Errorf("tag key %q contains prefix _ which is reserved", key)
		}
	}
	return nil
}

func (t Tags) Empty() bool {
	if t == nil {
		return true
	}
	return len(t) == 0
}

func (t Tags) Equals(t1 Tags) bool {
	if len(t) != len(t1) {
		return false
	}
	for key, val := range t {
		if t[key] != val {
			return false
		}
	}
	return true
}

func (t Tags) Contains(t1 Tags) bool {
	if len(t) < len(t1) {
		return false
	}
	count := 0
	for key, val := range t1 {
		if t[key] == val {
			count++
		} else {
			return false
		}
	}
	return count == len(t1)
}

func (t Tags) Has(key string) bool {
	_, ok := t[key]
	return ok
}

func (t Tags) Set(key string, value string) {
	t[key] = value
}

func (t Tags) Get(key string) string {
	return t[key]
}

func (t Tags) Diff(t1 Tags) (add Tags, del Tags, update Tags) {
	handled := make(map[string]bool, len(t)+len(t1))
	add, del, update = Tags{}, Tags{}, Tags{}
	for key, val := range t {
		handled[key] = true
		if val1, exist := t1[key]; exist {
			if val1 != val {
				update[key] = val1
			}
		} else {
			del[key] = val
		}
	}
	for key, val := range t1 {
		if !handled[key] {
			add[key] = val
		}
	}
	return
}

func (t Tags) ToJSON() string {
	data, _ := json.Marshal(t)
	return string(data)
}

type DataFrame struct {
	Data []byte

	Index Index

	Tags Tags

	Message string
}

func (f *DataFrame) Empty() bool {
	return len(f.Data) == 0 && len(f.Tags) == 0 && len(f.Message) == 0
}

type DataFrameEncoder interface {
	// Encode encodes DataFrame into bytes.
	Encode(frame DataFrame, w io.Writer) error

	// EncodeList attempts to encode batch of DataFrame into bytes.
	EncodeList(list []DataFrame, w io.Writer) error
}

type DataFrameDecoder interface {
	// Decode decodes DataFrame from bytes.
	Decode([]byte, *DataFrame) error

	// DecodeList attempts to decode DataFrameList from bytes.
	DecodeList([]byte) ([]DataFrame, error)
}

type AttributesEncoder interface {
	Encode(Attributes Attributes, w io.Writer) error
}

type AttributesDecoder interface {
	Decode([]byte, *Attributes) error
}

// DataFrameCodec helps to encode or decode a DataFrame from or to bytes.
type DataFrameCodec interface {
	MediaType() string

	DataFrameEncoder

	DataFrameDecoder
}

// AttributesCodec helps to encode or decode Attributes from or to bytes.
type AttributesCodec interface {
	MediaType() string

	AttributesEncoder

	AttributesDecoder
}

func LargestIndex(dfs []DataFrame) Index {
	var max Index
	for i := range dfs {
		if dfs[i].Index > max {
			max = dfs[i].Index
		}
	}
	return max
}

// Watcher is the entity following the stream.
type Watcher interface {
	// Watcher is a kind of DataFrameReader.
	DataFrameReader
	// Close stops Watcher and closes the FrameChan.
	Close()
}

type Attributes map[string]string

// const attributes keys the queue service implement must provide.
const (
	// const attributes keys the queue service implement must provide.
	Backend                    = "meta.backend"
	Name                       = "meta.name"
	State                      = "meta.state"
	MaxPayloadBytes            = "meta.maxPayloadBytes"
	UserIdentifyHeader         = "meta.header.user"
	GroupIdentifyHeader        = "meta.header.group"
	PriorityHeader             = "meta.header.priority"
	StreamLength               = "stream.length"
	StreamApproximateMaxLength = "stream.approxMaxLength"
	StreamLastEntry            = "stream.lastEntry"
	StreamFirstEntry           = "stream.firstEntry"

	// not necessary attribute keys
	ConsumersTotal = "consumers.status.total"
)

var MaxIndex = FromUint64(uint64(math.MaxUint64))

type Priority int64

func (p Priority) String() string {
	return strconv.FormatInt(int64(p), 10)
}

type Code string

func (c Code) String() string {
	return string(c)
}

const (
	codeValidRegexp = `^([A-Z][a-z]*)+(.[A-Z][a-z]*)?$|^[0-9]+$`
)

var codeRegexp = regexp.MustCompile(codeValidRegexp)

func (c Code) Verify() error {
	if codeRegexp.MatchString(c.String()) {
		if len(c) > 20 {
			return errors.New("code can not have more than 20 characters")
		}
		return nil
	} else {
		return errors.New("code must match regular expression: " + codeValidRegexp)
	}
}

// some well-known Codes.
const (
	// Shutdown indicates the consumer will exit soon.
	Shutdown Code = "Shutdown"
)

type DataFrameReader interface {
	// FrameChan return a DataFrame channel.
	FrameChan() <-chan DataFrame
}

// User authenticated information.
type User interface {
	// Uid represents the user id.
	Uid() string

	// Gid represents the group id of user.
	Gid() string

	// Token represents the access token of the queue service.
	Token() string
}

type UserAware interface {
	// User returns the user info.
	User() User
}

type UserWithToken interface {
	// Token to access the backend service.
	Token() string
}

const (
	userKey = "__user__"
)

// WithUser saves User into context.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext loads User from context.
func UserFromContext(ctx context.Context) (User, bool) {
	i := ctx.Value(userKey)
	if u, ok := i.(User); ok {
		return u, ok
	}
	return nil, false
}

type WorkerStatus string

const (
	WorkerRunning WorkerStatus = "Running"
	WorkerStopped WorkerStatus = "Stopped"
	WorkerError   WorkerStatus = "Error"
	WorkerUnknown WorkerStatus = "Unknown"
)

type StreamStatus string

const (
	StreamOk     StreamStatus = "OK"
	StreamCancel StreamStatus = "Cancel"
	StreamEnd    StreamStatus = "End"
)

const (
	OffsetEOS Offset = "eos"
)

type Offset string

func (o Offset) IsInf() bool {
	low := strings.ToLower(string(o))
	return low == "inf" || low == "+inf"
}

func (o Offset) Uint64() (uint64, bool) {
	u, err := strconv.ParseUint(string(o), 10, 64)
	if err != nil {
		return 0, false
	}
	return u, true
}

func Compare(o1, o2 Offset) (int, error) {
	uint1, ok1 := o1.Uint64()
	uint2, ok2 := o2.Uint64()
	switch {
	case ok1 && ok2:
		if uint1 > uint2 {
			return 1, nil
		} else if uint1 < uint2 {
			return -1, nil
		} else {
			return 0, nil
		}
	case ok1 && !ok2:
		if o2 == OffsetEOS || o2.IsInf() {
			return -1, nil
		} else {
			return -2, fmt.Errorf("unexpected offset: %v", o2)
		}
	case !ok1 && ok2:
		if o1 == OffsetEOS || o1.IsInf() {
			return 1, nil
		} else {
			return -2, fmt.Errorf("unexpected offset: %v", o1)
		}
	default:
		if o1 == OffsetEOS && o2 == OffsetEOS {
			return 0, nil
		}
		if o1.IsInf() && o2.IsInf() {
			return 0, nil
		}

		return -2, fmt.Errorf("unexpected compare: %s vs %s", o1, o2)
	}
}
