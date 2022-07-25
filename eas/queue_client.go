package eas

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pai-eas/eas-golang-sdk/eas/types"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	HeaderRequestId     = "X-Eas-Queueservice-Request-Id"
	HeaderAuthorization = "Authorization"

	DefaultGroupName = "eas"
)

type QueueUser struct {
	uid   string
	gid   string
	token string
}

func NewQueueUser(uid, gid, token string) QueueUser {
	return QueueUser{
		uid, gid, token,
	}
}

func (c QueueUser) Uid() string {
	return c.uid
}

func (c QueueUser) Gid() string {
	return c.gid
}

func (c QueueUser) Token() string {
	return c.token
}

// QueueClient is client of queue server, which also implements queue service interface
type QueueClient struct {
	// HTTP client.
	httpClient *http.Client
	// base url of queue server.
	baseUrl *url.URL

	user types.User

	WebsocketWatch bool

	once sync.Once
	attr types.Attributes
	// codecs for data frame and attributes.
	DCodec types.DataFrameCodec
	ACodec types.AttributesCodec
}

func NewQueueClient(endpoint, queueName, token string) (*QueueClient, error) {
	baseUrl := strings.Join([]string{endpoint, "api/predict", queueName}, "/")
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	if len(u.Scheme) == 0 {
		u.Scheme = "http"
	}
	uid := uuid.New().String()
	gid := DefaultGroupName
	cli := &QueueClient{
		baseUrl:        u,
		httpClient:     &http.Client{},
		user:           NewQueueUser(uid, gid, token),
		WebsocketWatch: true, // Watch through websocket by default
		DCodec:         types.DataFrameCodecFor(types.ContentTypeProtobuf),
		ACodec:         types.AttributesCodecFor(types.ContentTypeProtobuf),
	}

	return cli, nil
}

func must(fn func() error) {
	if err := fn(); err != nil {
		panic(err)
	}
}

func readMessage(reader io.Reader) string {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (q *QueueClient) getAttr(force bool) (types.Attributes, error) {
	if q.attr == nil {
		q.once.Do(func() { must(q.obtainAttr) })
	} else if force {
		if err := q.obtainAttr(); err != nil {
			return q.attr, fmt.Errorf("failed to obtain attributes, error: %v", err)
		}
	}

	return q.attr, nil
}

func (q *QueueClient) obtainAttr() error {
	// make a copy of base url.
	u := *q.baseUrl
	qe := u.Query()
	qe.Set("_attrs_", "true")
	u.RawQuery = qe.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	q.withAuthorization(req)
	req.Header.Set("accept", q.ACodec.MediaType())
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("visiting: %s, unexpected status code: %d, body: %s", u.String(), resp.StatusCode, string(body))
	}
	q.attr = types.Attributes{}
	if err = q.ACodec.Decode(body, &q.attr); err != nil {
		return err
	}
	return nil
}

// withIdentity populates user and group id into request.
func (q *QueueClient) withIdentity(req *http.Request) {
	attr, err := q.getAttr(false)
	if err != nil {
		// todo
	}
	uidHeader := attr[types.UserIdentifyHeader]
	gidHeader := attr[types.GroupIdentifyHeader]
	if len(uidHeader) == 0 {
		panic("malformed attributes")
	} else {
		req.Header.Add(uidHeader, q.user.Uid())
	}
	if len(gidHeader) > 0 {
		req.Header.Add(gidHeader, q.user.Gid())
	}
}

// withIdentity populates user and group id into request.
func (q *QueueClient) withAuthorization(req *http.Request) {
	if t, ok := q.user.(types.UserWithToken); ok {
		req.Header.Add(HeaderAuthorization, t.Token())
	}
}

func (q *QueueClient) Truncate(ctx context.Context, index uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_index_", strconv.FormatUint(index, 10))
	eq.Set("_trunc_", boolString(true))
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	q.withIdentity(req)
	q.withAuthorization(req)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) End(ctx context.Context, force bool) error {
	// make a copy of base url.
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_eos_", boolString(true))
	if force {
		eq.Set("_force_", boolString(true))
	}
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return err
	}
	q.withIdentity(req)
	q.withAuthorization(req)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) Put(ctx context.Context, data []byte, tags types.Tags) (index uint64, requestId string, err error) {
	// make a copy of base url.
	u := *q.baseUrl
	qe := u.Query()
	for key, val := range tags {
		qe.Set(key, val)
	}
	u.RawQuery = qe.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, requestId, err
	}
	q.withIdentity(req)
	q.withAuthorization(req)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return 0, requestId, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, requestId, err
	}
	requestId = resp.Header.Get(HeaderRequestId)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return 0, requestId, fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, string(body))
	}

	index, err = strconv.ParseUint(string(body), 0, 64)
	if err != nil {
		return 0, requestId, err
	}
	return index, requestId, nil
}

func (q *QueueClient) GetByIndex(ctx context.Context, index uint64) (dfs []types.DataFrame, err error) {
	return q.Get(ctx, index, 1, time.Duration(0), true, types.Tags{})
}

func (q *QueueClient) GetByRequestId(ctx context.Context, requestId string) (dfs []types.DataFrame, err error) {
	return q.Get(ctx, 0, 1, time.Duration(0), true, types.Tags{"requestId": requestId})
}

func (q *QueueClient) Get(ctx context.Context, index uint64, length int, timeout time.Duration, autoDelete bool, tags types.Tags) (dfs []types.DataFrame, err error) {
	var ret []types.DataFrame
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_index_", strconv.FormatUint(index, 10))
	eq.Set("_length_", strconv.FormatInt(int64(length), 10))
	eq.Set("_timeout_", timeout.String())
	eq.Set("_raw_", boolString(false))
	eq.Set("_auto_delete_", boolString(autoDelete))
	if err = tags.Validate(); err != nil {
		return nil, err
	}
	for key, val := range tags {
		eq.Set(key, val)
	}
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ret, err
	}
	req.Header.Set("Accept", q.DCodec.MediaType())
	q.withIdentity(req)
	q.withAuthorization(req)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return ret, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return ret, fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, string(data))
	}

	return q.DCodec.DecodeList(data)
}

func boolString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

type websocketWatch struct {
	ctx     context.Context
	cancel  context.CancelFunc
	conn    *websocket.Conn
	decoder types.DataFrameDecoder
	ch      chan types.DataFrame
}

func newWebsocketWatcher(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, decoder types.DataFrameDecoder) types.Watcher {
	w := &websocketWatch{
		ctx:     ctx,
		cancel:  cancel,
		conn:    conn,
		decoder: decoder,
		ch:      make(chan types.DataFrame, 100),
	}
	go w.run()
	return w
}

func (w *websocketWatch) FrameChan() <-chan types.DataFrame {
	return w.ch
}

func (w *websocketWatch) Close() {
	w.cancel()
}

func (w *websocketWatch) run() {
	go func() {
		<-w.ctx.Done()
		w.conn.Close()
	}()
	defer w.cancel()
	defer close(w.ch)
	var data []byte
	for {
		df := types.DataFrame{}
		err := websocket.Message.Receive(w.conn, &data)
		if err != nil {
			df.Message = fmt.Sprintf("error reading message: %v", err)
			w.ch <- df
			return
		}
		if err = w.decoder.Decode(data, &df); err != nil {
			df.Message = fmt.Sprintf("failed to decode message: %v", err)
		}
		w.ch <- df
	}
}

type httpWatcher struct {
	ctx     context.Context
	cancel  context.CancelFunc
	reader  io.ReadCloser
	decoder types.DataFrameDecoder
	ch      chan types.DataFrame
}

func newHTTPWatcher(ctx context.Context, cancel context.CancelFunc, reader io.ReadCloser, decoder types.DataFrameDecoder) *httpWatcher {
	w := &httpWatcher{
		ctx:     ctx,
		cancel:  cancel,
		reader:  reader,
		decoder: decoder,
		ch:      make(chan types.DataFrame, 100),
	}
	go w.run()
	return w
}

func (h *httpWatcher) FrameChan() <-chan types.DataFrame {
	return h.ch
}

func (h *httpWatcher) Close() {
	h.cancel()
	h.reader.Close()
}

func (h *httpWatcher) run() {
	go func() {
		<-h.ctx.Done()
		h.reader.Close()
	}()

	defer h.cancel()
	defer close(h.ch)
	rbuf := [4096]byte{}
	buf := bytes.NewBuffer(nil)
	for {
		n, err := h.reader.Read(rbuf[:])
		if n > 0 {
			io.Copy(buf, bytes.NewBuffer(rbuf[:n]))
			if err == io.ErrShortBuffer {
				continue
			} else if err != nil {
				// fmt.Printf("failed to read: %v\n", err)
				return
			}
			df := types.DataFrame{}
			if err = h.decoder.Decode(buf.Bytes(), &df); err != nil {
				// klog.Errorf("failed to decode, err: %v", err)
				return
			}
			buf.Reset()
			h.ch <- df
		} else {
			break
		}
	}
}

func (q *QueueClient) Watch(ctx context.Context, index, window uint64, indexOnly bool, autocommit bool) (types.Watcher, error) {
	ctx, cancel := context.WithCancel(ctx)
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_index_", strconv.FormatUint(index, 10))
	eq.Set("_window_", strconv.FormatUint(window, 10))
	eq.Set("_index_only_", boolString(indexOnly))
	eq.Set("_auto_commit_", boolString(autocommit))
	eq.Set("_watch_", boolString(true))
	u.RawQuery = eq.Encode()
	if q.WebsocketWatch {
		// use websocket watch.
		u.Scheme = "ws"
		config, err := websocket.NewConfig(u.String(), q.baseUrl.String())
		if err != nil {
			cancel()
			return nil, err
		}
		header := http.Header{}
		attr, err := q.getAttr(false)
		if err != nil {
			return nil, err
		}
		uidHeader := attr[types.UserIdentifyHeader]
		gidHeader := attr[types.GroupIdentifyHeader]
		// set websocket request headers.
		header.Set(uidHeader, q.user.Uid())
		header.Set("Accept", q.DCodec.MediaType())
		header.Set(HeaderAuthorization, q.user.Token())
		if len(gidHeader) > 0 {
			header.Set(gidHeader, q.user.Gid())
		}
		config.Header = header
		conn, err := websocket.DialConfig(config)
		if err != nil {
			cancel()
			return nil, err
		}
		return newWebsocketWatcher(ctx, cancel, conn, q.DCodec), nil

	} else {
		// default http watch.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			cancel()
			return nil, err
		}
		req.Header.Set("Accept", q.DCodec.MediaType())
		q.withIdentity(req)
		q.withAuthorization(req)
		resp, err := q.httpClient.Do(req)
		if err != nil {
			cancel()
			return nil, err
		}
		if resp.StatusCode != 200 {
			cancel()
			content, _ := ioutil.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, string(content))
		}
		reader := types.NewLengthDelimitedFrameReader(resp.Body)
		return newHTTPWatcher(ctx, cancel, reader, q.DCodec), nil
	}
}

func (q *QueueClient) Commit(ctx context.Context, indexes ...uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	var indexStr []string
	for _, idx := range indexes {
		indexStr = append(indexStr, strconv.FormatUint(idx, 10))
	}
	eq := u.Query()
	eq.Set("_indexes_", strings.Join(indexStr, ","))
	// eq.Set("_delete_", boolString(del))
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), nil)
	if err != nil {
		return err
	}
	q.withIdentity(req)
	q.withAuthorization(req)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) Del(ctx context.Context, indexes ...uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	var indexStr []string
	for _, idx := range indexes {
		indexStr = append(indexStr, strconv.FormatUint(idx, 10))
	}
	eq := u.Query()
	eq.Set("_indexes_", strings.Join(indexStr, ","))
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	q.withIdentity(req)
	q.withAuthorization(req)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) Attributes() (types.Attributes, error) {
	return q.getAttr(true)
}
